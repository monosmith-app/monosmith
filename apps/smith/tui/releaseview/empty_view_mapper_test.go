package releaseview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func emptyLines(m Model) ([]string, map[int]int) {
	return m.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 120)
}

func TestAnEmptyPlanSaysOneHeadlineAndFoldsTheExplanation(t *testing.T) {
	m := FromResult(Result{Plan: release.Plan{Unchanged: []string{"sdk"}}})
	out := render(m)
	if !strings.Contains(out, "Nothing to release: no feat: or fix: commits since the last tag") || !strings.Contains(out, "> Why? (w)") {
		t.Fatalf("headline or fold missing:\n%s", out)
	}
	if strings.Contains(out, "Only feat:") || strings.Contains(out, "Create the first release commit") {
		t.Fatalf("the explanation is not folded:\n%s", out)
	}
	if n := strings.Count(out, "Nothing to release"); n != 1 {
		t.Fatalf("the headline is said %d times:\n%s", n, out)
	}
	out = render(m.ToggleWhy())
	for _, want := range []string{"v Why? (w)", "Only feat: (a minor version) and fix: (a patch)", "A chore: commit releases nothing"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

func TestTheHeadlineNamesAHeldBackPackageOrAWorkspaceWithoutPackages(t *testing.T) {
	held := FromResult(Result{Plan: release.Plan{Held: []string{"kit"}, Unchanged: []string{"web"}}})
	if out := render(held); !strings.Contains(out, "you held back kit") {
		t.Fatalf("held:\n%s", out)
	}
	none := FromResult(Result{})
	if out := render(none); !strings.Contains(out, "this workspace has no package") {
		t.Fatalf("no package:\n%s", out)
	}
}

func TestOnlyAFirstReleaseOffersTheFirstCommitAndItsRowIsAHit(t *testing.T) {
	first := FromResult(Result{Plan: release.Plan{Unchanged: []string{"lib"}, NeverReleased: []string{"lib"}}})
	first.FirstButton = "[ Create the first release commit (c) ]"
	lines, rowOf := emptyLines(first)
	out := strings.Join(lines, "\n")
	if !strings.Contains(out, "Create the first release commit (c)") || !strings.Contains(out, `$ git commit --allow-empty -m "feat: first version"`) {
		t.Fatalf("no first release action:\n%s", out)
	}
	rows := map[int]bool{}
	for _, v := range rowOf {
		rows[v] = true
	}
	if !rows[RowFirst] || !rows[RowWhy] {
		t.Fatalf("the action and the fold are no hits: %v", rowOf)
	}
	later := FromResult(Result{Plan: release.Plan{Unchanged: []string{"lib"}}})
	later.FirstButton = first.FirstButton
	if out := render(later); strings.Contains(out, "Create the first release commit") {
		t.Fatalf("a package that was released offers a first release:\n%s", out)
	}
}

func TestAPlanWithPackagesHasTheReleaseButtonBelowTheHeadline(t *testing.T) {
	m := planned()
	m.Button = "[ Release now (x) ]"
	lines, rowOf := emptyLines(m)
	if !strings.Contains(lines[0], "2 packages to release") || strings.TrimSpace(lines[1]) != m.Button || rowOf[1] != RowRelease {
		t.Fatalf("the button is not under the headline: %q %v", lines[:3], rowOf)
	}
}
