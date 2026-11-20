package release

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

func TestPromoteComparesTheStablePlanWithTheChannel(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	channels := []workspace.ReleaseChannel{{Branch: "dev", Prerelease: "beta"}}
	commit(t, g, "packages/core/a", "feat(core): first")
	for _, tag := range []string{"core@6.1.0", "sdk@1.0.0", "web@1.0.0"} {
		if err := g.CreateTag(tag, tag); err != nil {
			t.Fatal(err)
		}
	}
	commit(t, g, "packages/core/b", "feat(core): tested on dev")
	if err := g.CreateTag("core@6.2.0-beta.1", "core@6.2.0-beta.1"); err != nil {
		t.Fatal(err)
	}

	plan, err := MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	if core := byProject(plan)["core"]; core.Next != "6.2.0" {
		t.Fatalf("the stable plan releases what dev tested: %+v", core)
	}
	got := strings.Join(promoteProblems(g, pg, plan, channels), "\n")
	if strings.Contains(got, "core:") {
		t.Errorf("6.2.0 after 6.2.0-beta.1 is the tested version:\n%s", got)
	}
	if !strings.Contains(got, "sdk: 1.0.1 was never released on a channel (dev)") {
		t.Errorf("a project the channel never released is named:\n%s", got)
	}

	commit(t, g, "packages/core/c", "feat(core)!: straight to the release branch")
	plan, err = MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	got = strings.Join(promoteProblems(g, pg, plan, channels), "\n")
	if !strings.Contains(got, "core: this release plans 7.0.0, but the dev channel last released 6.2.0-beta.1") {
		t.Errorf("a plan above the tested version is named:\n%s", got)
	}

	if problems := promoteProblems(g, pg, plan, []workspace.ReleaseChannel{{Branch: "next", Prerelease: "beta", Major: 8}}); problems != nil {
		t.Errorf("a channel with a major is a line of its own and is not compared: %v", problems)
	}
}
