package release

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// onBranch runs git checkout for a test.
func onBranch(t *testing.T, g *gitclient.Git, args ...string) {
	t.Helper()
	if _, err := g.Run(append([]string{"checkout", "-q"}, args...)...); err != nil {
		t.Fatal(err)
	}
}

func tag(t *testing.T, g *gitclient.Git, name string) {
	t.Helper()
	if err := g.CreateTag(name, name); err != nil {
		t.Fatal(err)
	}
}

// previewSession is a repository whose release branch main has core@1.0.0, with
// the branch dev started from it and checked out.
func previewSession(t *testing.T) (*session.Session, *gitclient.Git) {
	t.Helper()
	g := repo(t)
	commit(t, g, "packages/core/a", "feat(core): first")
	tag(t, g, "core@1.0.0")
	onBranch(t, g, "-b", "dev")
	return &session.Session{Git: g, Graph: packages(t), Config: workspace.Config{Release: workspace.ReleaseSection{Branch: "main"}}}, g
}

func TestAPreviewPlansWhatMergingTheBranchWouldRelease(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/core/b", "feat(core): on dev")
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	core := byProject(plan)["core"]
	if plan.PreviewOf != "main" || core.Current != "1.0.0" || core.Next != "1.1.0" || core.Tag != "core@1.1.0" {
		t.Fatalf("plan %+v", plan)
	}
	if got := PreviewNotice("dev", plan); got != "Preview: what merging dev into main would release" {
		t.Fatalf("notice %q", got)
	}
}

func TestAPreviewOfABranchBehindTheReleaseBranchStartsFromItsNewerTag(t *testing.T) {
	s, g := previewSession(t)
	onBranch(t, g, "main")
	commit(t, g, "packages/core/m", "feat(core): released on main")
	tag(t, g, "core@1.1.0")
	onBranch(t, g, "dev")
	commit(t, g, "packages/core/d", "fix(core): on dev")
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	core := byProject(plan)["core"]
	if core.Current != "1.1.0" || core.Next != "1.1.1" || !reflect.DeepEqual(core.Commits, []string{"fix(core): on dev"}) {
		t.Fatalf("core %+v", core)
	}
}

func TestAPreviewKeepsTheUnreleasedFixOfTheReleaseBranch(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/core/d", "chore(core): only a chore on dev")
	onBranch(t, g, "main")
	commit(t, g, "packages/core/m", "fix(core): unreleased on main")
	onBranch(t, g, "dev")
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	core := byProject(plan)["core"]
	if core.Current != "1.0.0" || core.Next != "1.0.1" || !reflect.DeepEqual(core.Commits, []string{"fix(core): unreleased on main"}) {
		t.Fatalf("core %+v", core)
	}
}

func TestAPreviewOfAProjectNeverReleasedCountsBothBranches(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/gomod/d", "fix: on dev")
	onBranch(t, g, "main")
	commit(t, g, "packages/gomod/m", "feat: on main")
	onBranch(t, g, "dev")
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	gomod := byProject(plan)["gomod"]
	if gomod.Current != "" || gomod.Next != "0.1.0" || len(gomod.Commits) != 2 {
		t.Fatalf("gomod %+v", gomod)
	}
}

func TestAMissingReleaseBranchGivesTheOrdinaryPlan(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/core/b", "feat(core): on dev")
	s.Config.Release.Branch = "master"
	got, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	want, err := MakePlan(g, s.Graph)
	if err != nil {
		t.Fatal(err)
	}
	if got.PreviewOf != "" || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestAPreviewUsesTheRemoteBranchWhenTheLocalOneIsMissing(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/core/b", "feat(core): on dev")
	if _, err := g.Run("update-ref", "refs/remotes/origin/main", "main"); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run("branch", "-q", "-m", "main", "trunk"); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PreviewOf != "main" || byProject(plan)["core"].Next != "1.1.0" {
		t.Fatalf("plan %+v", plan)
	}
}

func TestAPrefixWithTagsOnlyOnTheBranchKeepsThoseTagsAsItsLastRelease(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/gomod/d", "feat: on dev")
	tag(t, g, "packages/gomod/v1.0.0")
	onBranch(t, g, "main")
	commit(t, g, "packages/gomod/m", "fix: on main")
	onBranch(t, g, "dev")
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	gomod := byProject(plan)["gomod"]
	if gomod.Current != "1.0.0" || gomod.Next != "1.0.1" || !reflect.DeepEqual(gomod.Commits, []string{"fix: on main"}) {
		t.Fatalf("gomod %+v", gomod)
	}
}

func TestOnTheReleaseBranchThePlanIsTheOrdinaryOne(t *testing.T) {
	s, g := previewSession(t)
	onBranch(t, g, "main")
	commit(t, g, "packages/core/m", "fix(core): on main")
	got, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	want, err := MakePlan(g, s.Graph)
	if err != nil {
		t.Fatal(err)
	}
	if got.PreviewOf != "" || len(got.Items) == 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if PreviewNotice("main", got) != "" {
		t.Fatal("a plan that is not a preview has no notice")
	}
}

func TestAChannelBranchAndADetachedCheckoutAreNoPreview(t *testing.T) {
	s, g := previewSession(t)
	commit(t, g, "packages/core/b", "feat(core): on dev")
	s.Config.Release.Channels = []workspace.ReleaseChannel{{Branch: "dev", Prerelease: "beta", Major: 2}}
	plan, err := PlanFor(s)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PreviewOf != "" || plan.Channel != "beta" {
		t.Fatalf("a channel branch plans its channel: %+v", plan)
	}
	s.Config.Release.Channels = nil
	onBranch(t, g, "--detach")
	if ref, _ := PreviewBase(g, s.Config.Release); ref != "" {
		t.Fatalf("a detached checkout has no preview, got %q", ref)
	}
}
