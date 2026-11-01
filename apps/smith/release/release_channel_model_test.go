package release

import (
	"strings"
	"testing"

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
	pg.Get("core").Release = &workspace.ProjectRelease{NPM: []string{"@acme/core"}}
	plan := Plan{Items: []Item{{Project: "core", Plugin: "test"}, {Project: "sdk", Plugin: "other"}}}
	got := strings.Join(channelProblems(plan, pg, func(p string) bool { return p == "test" }), "\n")
	for _, want := range []string{
		"sdk: the other plugin cannot publish a release channel",
		"sdk: docker deploy a release channel",
		"core: the binaries sent to package managers and registries have no release channel yet",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "core: the test plugin") {
		t.Errorf("a plugin with channels is accepted:\n%s", got)
	}
}
