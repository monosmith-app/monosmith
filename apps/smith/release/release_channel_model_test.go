package release

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestPrereleaseCountersOrderByValue(t *testing.T) {
	beta9, _ := ParseVersion("2.0.0-beta.9")
	beta10, _ := ParseVersion("2.0.0-beta.10")
	rc1, _ := ParseVersion("2.0.0-rc.1")
	stable, _ := ParseVersion("2.0.0")
	if !beta9.Less(beta10) || beta10.Less(beta9) {
		t.Fatal("beta.9 must come before beta.10")
	}
	if !beta10.Less(rc1) || !rc1.Less(stable) {
		t.Fatal("beta.10 < rc.1 < 2.0.0")
	}
}

func TestAChannelBranchReleasesItsOwnLineAndLeavesTheStableLineAlone(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	ch := &Channel{ID: "beta", Major: 2}
	commit(t, g, "packages/core/a", "feat(core): first")
	if err := g.CreateTag("core@1.4.0", "core@1.4.0"); err != nil {
		t.Fatal(err)
	}
	stable := g.CurrentBranch()
	if _, err := g.Run("checkout", "-q", "-b", "next"); err != nil {
		t.Fatal(err)
	}
	releaseNext := func(wantNext, wantCurrent string) {
		t.Helper()
		plan, err := MakeChannelPlan(g, pg, ch)
		if err != nil {
			t.Fatal(err)
		}
		core := byProject(plan)["core"]
		if core.Next != wantNext || core.Current != wantCurrent || core.Tag != "core@"+wantNext {
			t.Fatalf("core %+v, want %s after %q", core, wantNext, wantCurrent)
		}
		if err := g.CreateTag(core.Tag, core.Tag); err != nil {
			t.Fatal(err)
		}
	}

	commit(t, g, "packages/core/b", "feat(core): the first of the next line")
	releaseNext("2.0.0-beta.1", "1.4.0")
	commit(t, g, "packages/core/c", "fix(core)!: a breaking fix stays inside the channel")
	releaseNext("2.0.0-beta.2", "2.0.0-beta.1")
	for _, tag := range []string{"core@2.0.0-beta.9"} {
		if err := g.CreateTag(tag, tag); err != nil {
			t.Fatal(err)
		}
	}
	commit(t, g, "packages/core/d", "fix(core): counts by value")
	releaseNext("2.0.0-beta.10", "2.0.0-beta.9")

	commit(t, g, "README.md", "docs: nothing to release")
	plan, err := MakeChannelPlan(g, pg, ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 0 {
		t.Fatalf("a docs commit releases nothing on a channel: %+v", plan.Items)
	}

	plan, err = MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	if core := byProject(plan)["core"]; core.Current != "1.4.0" {
		t.Fatalf("the stable plan ignores channel tags: %+v", core)
	}

	if _, err := g.Run("checkout", "-q", stable); err != nil {
		t.Fatal(err)
	}
	commit(t, g, "packages/core/e", "fix(core): a stable fix")
	plan, err = MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	if core := byProject(plan)["core"]; core.Next != "1.4.1" {
		t.Fatalf("the stable line keeps releasing 1.x: %+v", core)
	}
	channelPlan, err := MakeChannelPlan(g, pg, ch)
	if err != nil {
		t.Fatal(err)
	}
	if core := byProject(channelPlan)["core"]; core.Next != "2.0.0-beta.1" {
		t.Fatalf("a channel with no release reachable starts again: %+v", core)
	}
}

func TestBranchProblem(t *testing.T) {
	cfg := workspace.ReleaseSection{Branch: "main", Channels: []workspace.ReleaseChannel{{Branch: "next", Prerelease: "beta", Major: 2}}}
	for branch, want := range map[string]string{
		"main":    "",
		"":        "",
		"feature": "releases run from main, not feature",
		"next":    "",
	} {
		if got := branchProblem(cfg, branch); got != want {
			t.Errorf("branch %q: got %q, want %q", branch, got, want)
		}
	}
}

func TestAChannelPublishesUnderItsDistTagOrItsPrerelease(t *testing.T) {
	cfg := workspace.ReleaseSection{Branch: "main", Channels: []workspace.ReleaseChannel{
		{Branch: "next", Prerelease: "beta", Major: 2},
		{Branch: "edge", Prerelease: "rc", Major: 3, DistTag: "edge"},
	}}
	if got := ChannelOf(cfg, "next").Tag; got != "beta" {
		t.Errorf("default tag %q", got)
	}
	if got := ChannelOf(cfg, "edge").Tag; got != "edge" {
		t.Errorf("dist_tag %q", got)
	}
}

func TestAChannelReleaseNeedsAPluginThatPublishesChannels(t *testing.T) {
	pg := packages(t)
	pg.Get("sdk").Deployers = []string{"docker"}
	pg.Get("core").Release = &workspace.ProjectRelease{NPM: []string{"@acme/core"}, PackageManagers: []string{"brew"}}
	pg.Get("web").Release = &workspace.ProjectRelease{NPM: []string{"@acme/web"}, PyPI: "acme-web", Pub: "acme_web"}
	plan := Plan{Items: []Item{{Project: "core", Plugin: "test"}, {Project: "sdk", Plugin: "other"}, {Project: "web", Plugin: "other"}}}
	got := strings.Join(channelProblems(plan, pg, func(p string) bool { return p == "test" }), "\n")
	for _, want := range []string{
		"sdk: the other plugin cannot publish a release channel",
		"channels = false in [projects.sdk.release]",
		"sdk: docker deploy a release channel",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	for _, not := range []string{"core: the test plugin", "core: the package", "web:"} {
		if strings.Contains(got, not) {
			t.Errorf("%q must be accepted: a plugin with channels, and an app whose binaries go to registries that order prereleases below releases:\n%s", not, got)
		}
	}
}

func TestAProjectWithChannelsFalseIsLeftOutOfAChannelPlan(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	no := false
	pg.Get("gomod").Release = &workspace.ProjectRelease{Channels: &no}
	commit(t, g, "packages/core/a", "feat(core): in")
	commit(t, g, "packages/gomod/a", "feat: out")
	plan, err := MakeChannelPlan(g, pg, &Channel{ID: "beta", Major: 2, Tag: "beta"})
	if err != nil {
		t.Fatal(err)
	}
	items := byProject(plan)
	if _, ok := items["gomod"]; ok || items["core"].Next != "2.0.0-beta.1" {
		t.Fatalf("plan %+v", plan.Items)
	}
	if stable, _ := MakePlan(g, pg); byProject(stable)["gomod"].Next != "0.1.0" {
		t.Fatal("the stable line still releases the project")
	}
}

func TestAPlanAndItsPreviewNameTheirChannel(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	commit(t, g, "packages/core/a", "feat(core): first")
	plan, err := MakeChannelPlan(g, pg, &Channel{ID: "beta", Major: 2, Tag: "next"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Channel != "next" {
		t.Fatalf("channel %q", plan.Channel)
	}
	if md := (Report{DryRun: true, Channel: plan.Channel}).Markdown(); !strings.Contains(md, "Release channel `next`: prerelease versions, published under the `next` tag") {
		t.Errorf("the preview must name the channel:\n%s", md)
	}
	if stable, _ := MakePlan(g, pg); stable.Channel != "" || strings.Contains((Report{DryRun: true}).Markdown(), "channel") {
		t.Error("a stable plan names no channel")
	}
}

func TestAPullRequestBuildPlansForTheBranchItTargets(t *testing.T) {
	g := repo(t)
	commit(t, g, "README.md", "docs: start")
	s := &session.Session{Git: g}
	if got := PlanBranch(s); got != g.CurrentBranch() || got == "" {
		t.Fatalf("on a branch the plan is for that branch, got %q", got)
	}
	if _, err := g.Run("checkout", "-q", "--detach"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_BASE_REF", "next")
	if got := PlanBranch(s); got != "next" {
		t.Errorf("GitHub: %q", got)
	}
	t.Setenv("GITHUB_BASE_REF", "")
	t.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "refs/heads/next")
	if got := PlanBranch(s); got != "next" {
		t.Errorf("Azure: %q", got)
	}
}
