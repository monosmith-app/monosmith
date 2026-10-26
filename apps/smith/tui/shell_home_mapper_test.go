package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAWorkspaceOpensOnHomeAndChecksItselfInTheBackground(t *testing.T) {
	sh := NewShell(Config{Theme: theme.Default(), Sources: &sources.Sources{Doctor: func() doctorview.Result { return doctorview.Result{} }}})
	if sh.tab != commands.TabHome {
		t.Fatalf("opens on %q", sh.tab)
	}
	if !sh.doctor.Loading || sh.Init() == nil {
		t.Fatal("doctor runs from the start, before anyone opens the Doctor view")
	}
	if !strings.Contains(pane.Strip(sh.healthSummary()), "checking") {
		t.Fatalf("the header says it is checking: %q", sh.healthSummary())
	}
}

func TestHomeLeadsWithTheHealthAndOneNextActionWithItsKey(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	// The header badge is filled without the Doctor view having been opened.
	if first := l.lines()[0]; !strings.Contains(first, "✗2") {
		t.Fatalf("the header carries the health: %q", first)
	}
	l.has("2 checks fail", "Next: 2 checks fail, 1 fix is ready: Enter opens the fix wizard", "1 of 4 projects changed", "nothing has run yet", "not checked: 5 opens Deps")
	lines := l.lines()
	var headline, next int
	for i, line := range lines {
		if strings.Contains(line, "2 checks fail") && headline == 0 {
			headline = i
		}
		if strings.Contains(line, "Next:") {
			next = i
		}
	}
	if headline == 0 || next <= headline {
		t.Fatalf("the headline (row %d) comes before the next action (row %d)", headline, next)
	}
	footer := lines[len(lines)-2]
	for _, want := range []string{"Enter next", "b build", "w changed", "d doctor", "1-7 views"} {
		if !strings.Contains(footer, want) {
			t.Errorf("the Home footer lacks %q: %q", want, footer)
		}
	}
	if strings.Contains(footer, "f follow") || strings.Contains(footer, "x fix all") {
		t.Fatalf("keys of other views on Home: %q", footer)
	}
}

func TestEnterAndAClickDoTheSuggestedAction(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	// With a fix ready the suggestion opens the fix wizard, not the Doctor
	// tab and not a non-interactive `doctor --fix`.
	l.keys("enter")
	if l.sh().fix == nil || l.sh().tab != commands.TabHome || len(l.started) != 0 {
		t.Fatalf("Enter did the next action: wizard %v, tab %q, started %v", l.sh().fix != nil, l.sh().tab, l.started)
	}
	l.has("Fix · review", "[ Apply  Enter ]")
	l = newHomeLab(t, 140, 34)
	l.clickText("Next: 2 checks fail")
	if l.sh().fix == nil {
		t.Fatal("a click on the next action did not open the fix wizard")
	}
	// Without a fix doctor can make, it goes to the Doctor tab as before.
	l = newHomeLab(t, 140, 34)
	sh := l.sh()
	sh.doctor.Report.Findings[1].Fix, sh.doctor.Report.Fixes = nil, nil
	l.m = sh
	l.has("Next: 2 checks fail: press d")
	l.keys("enter")
	if l.sh().tab != commands.TabDoctor || l.sh().fix != nil {
		t.Fatalf("with nothing to fix Enter opened %q", l.sh().tab)
	}
}

func TestHomeKeysDoWhatTheirHintsSay(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.keys("w")
	if l.lastStarted() != "run lint test --affected" {
		t.Fatalf("w started %q", l.lastStarted())
	}
}

func healthyShell(t *testing.T, affected bool) *lab {
	t.Helper()
	l := newHomeLab(t, 140, 34)
	sh := l.sh()
	projects := sampleProjects
	if !affected {
		projects = nil
		for _, p := range sampleProjects {
			p.Affected = false
			projects = append(projects, p)
		}
	}
	sh.sources.Projects = func() sources.ProjectsLoaded { return sources.ProjectsLoaded{Projects: projects, Branch: "main"} }
	sh.sources.Doctor = func() doctorview.Result {
		return doctorview.Result{Report: doctor.Report{Passed: 3, Findings: []doctor.Finding{{Finding: protocol.Finding{Check: "git is installed", OK: true}}}}}
	}
	l.m = sh
	l.keys("r")
	return l
}

func TestWhenAllIsWellHomeSuggestsABuild(t *testing.T) {
	l := healthyShell(t, false)
	l.has("The workspace is healthy", "Next: all good: press b to build")
	if first := l.lines()[0]; strings.Contains(first, "✗") || !strings.Contains(first, "✓") {
		t.Fatalf("the header shows a clean bill: %q", first)
	}
}

func TestAffectedProjectsComeBeforeABuild(t *testing.T) {
	l := healthyShell(t, true)
	l.has("Next: 1 project affected: press w to lint and test them")
}

func TestOutdatedDependenciesAreCountedOnceTheyAreKnown(t *testing.T) {
	l := healthyShell(t, false)
	l.keys("5", "1")
	l.has("3 with newer releases", "Next: 3 dependencies have newer releases: press 5")
	sh := l.sh()
	sh.sources.Deps = func() depsview.Result { return depsview.Result{} }
	l.m = sh
	l.keys("5", "r", "1")
	l.has("every dependency is on its latest release", "all good")
}

func TestHomeTellsHowTheLastRunEnded(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("b")
	l.keys("1")
	l.has("Last run", "passed  run build -p orders")

	l = newLab(t, true, 140, 30)
	l.next = failingRun()
	l.keys("b")
	l.keys("1")
	l.has("failed  run build -p orders")
}

func TestHomeFitsTheLeastWidthAndReloadChecksAgain(t *testing.T) {
	l := newHomeLab(t, 72, 24)
	l.has("Next:", "Health")
	for i, line := range l.lines() {
		if w := len([]rune(line)); w > 72 {
			t.Fatalf("line %d is %d cells: %q", i, w, line)
		}
	}
	calls := 0
	sh := l.sh()
	inner := sh.sources.Doctor
	sh.sources.Doctor = func() doctorview.Result { calls++; return inner() }
	l.m = sh
	l.keys("r")
	if calls != 1 {
		t.Fatalf("r on Home checked %d times", calls)
	}
}
