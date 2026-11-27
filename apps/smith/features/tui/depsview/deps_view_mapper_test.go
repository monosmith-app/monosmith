package depsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

// detailText is the detail of the selected package as plain text: each
// heading, then its items.
func detailText(m Model, w int) string {
	var out []string
	for _, s := range m.Detail(theme.Default(), w, func(p string) string { return "apps/" + p }) {
		out = append(out, s.Heading)
		out = append(out, s.Items...)
	}
	return pane.Strip(strings.Join(out, "\n"))
}

func TestEachGroupHeadingCountsItsReleases(t *testing.T) {
	text := strings.Join(render(sample(), 120, 60), "\n")
	for _, want := range []string{"Patch and minor releases · 18 · safe to take together", "Major releases · 4 · may break your code"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "↑ 22 with newer releases · 12 patch · 6 minor · 4 major") {
		t.Fatalf("the headline counts each level:\n%s", text)
	}
}

func TestAChipCountsOnlyTheGroupItLetsThrough(t *testing.T) {
	m := sample().SetFilter(FilterMajor)
	text := strings.Join(render(m, 120, 60), "\n")
	if strings.Contains(text, "Patch and minor releases ·") || !strings.Contains(text, "Major releases · 4 ·") {
		t.Fatalf("the Major chip shows the major group alone:\n%s", text)
	}
}

func TestEachGroupHasItsColumnHeadingAndARule(t *testing.T) {
	lines, rowOf := sample().View(theme.Default(), 120, 60)
	headings := 0
	for i, l := range lines {
		if r, ok := rowOf[i]; ok && r == HeadingRow && strings.Contains(pane.Strip(l), "Package") {
			headings++
			if !strings.Contains(pane.Strip(lines[i+1]), "────") {
				t.Fatalf("no rule under the heading: %q", pane.Strip(lines[i+1]))
			}
		}
	}
	if headings != 2 {
		t.Fatalf("%d column headings, want one per group", headings)
	}
}

func TestTheDetailTellsThePackage(t *testing.T) {
	m := FromResult(Result{Rows: []Row{{Name: "react", Current: "17.0.2", Latest: "18.3.0", Level: "major", Kinds: []string{"prod"}, Projects: []string{"web", "kit"}}}})
	text := detailText(m, 40)
	for _, want := range []string{"17.0.2 → 18.3.0  ▲ major", "can break your code", "Used by", "web  apps/web", "kit  apps/kit", "Declared as", "prod", "The same as", "$ smith deps up react --major", "One version"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the detail lacks %q:\n%s", want, text)
		}
	}
	if m.DetailTitle() != "react" {
		t.Fatalf("title %q", m.DetailTitle())
	}
}

func TestWithNoPackageTheDetailIsThePolicy(t *testing.T) {
	m := FromResult(Result{})
	got := m.Detail(theme.Default(), 40, nil)
	if len(got) != 1 || got[0].Heading != PolicyTitle() || m.DetailTitle() != PolicyTitle() {
		t.Fatalf("detail %+v", got)
	}
}

func TestPackageArgvAsksForMajorOnlyForAMajorRelease(t *testing.T) {
	if got := strings.Join(PackageArgv(Row{Name: "zod", Level: "minor"}), " "); got != "deps up zod" {
		t.Errorf("minor: %q", got)
	}
	if got := strings.Join(PackageArgv(Row{Name: "react", Level: "major"}), " "); got != "deps up react --major" {
		t.Errorf("major: %q", got)
	}
}

func TestManyProjectsShrinkToTheFirstAndACount(t *testing.T) {
	if got := usedText([]string{"orders-api", "core", "web"}, 12); got != "orders-api +2" {
		t.Errorf("got %q", got)
	}
	if got := usedText([]string{"web", "kit"}, 12); got != "web, kit" {
		t.Errorf("got %q", got)
	}
}
