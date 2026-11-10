package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/composer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

func TestTheTabBarNamesTheAddFormWhileItsConfirmationIsOpen(t *testing.T) {
	argv := []string{"add", "app", "web", "--lang", "node", "--with", "react", "--dry-run"}
	l := newLab(t, true, 110, 32)
	sh := l.sh()
	sh.tab, sh.Form = commands.TabGuide, "Add a project"
	sh.addPlanRead(capturedMsg{argv: argv, res: sources.Captured{OK: true, Text: realAddPlan}})
	l.m = sh
	shown := l.sh()
	if got := shown.page().Title(); got != "Add a project" {
		t.Fatalf("the right label of the tab bar says %q", got)
	}
	if head := strings.Join(l.lines()[:5], "\n"); strings.Contains(head, "Start") {
		t.Fatalf("the tab bar still says Start:\n%s", head)
	}
}

func TestTheStartTabKeepsItsNameWhenAConfirmationBelongsToAnotherTab(t *testing.T) {
	l := newLab(t, true, 110, 32)
	sh := l.sh()
	sh.tab, sh.Form = commands.TabRelease, "Add a project"
	sh.compose = &composer.Model{}
	if got := sh.guideModalName(); got != "" {
		t.Fatalf("another tab is open, but the Start tab is called %q", got)
	}
}

// stripText is the tab strip as plain text.
func stripText(l *lab) string {
	var rows []string
	for _, line := range l.sh().tabStrip().Lines {
		rows = append(rows, pane.Strip(line))
	}
	return strings.Join(rows, "\n")
}

func TestTheTabNamesDoNotChangeWhenABadgeAppears(t *testing.T) {
	spelled := func(text string) string {
		var out []string
		for _, name := range []string{"Projects", "Proj ", "Release", "Rel "} {
			if strings.Contains(text, name) {
				out = append(out, name)
			}
		}
		return strings.Join(out, "|")
	}
	for w := 72; w <= 150; w++ {
		l := newLab(t, true, w, 30)
		l.next = &scripted{lines: []string{"kit:build × exited with code 2"}, exit: runner.Exit{Code: 1}}
		before := spelled(stripText(l))
		l.keys("b")
		l.advance(time.Second)
		l.keys("2")
		if after := spelled(stripText(l)); after != before {
			t.Errorf("%d columns: tab names changed from %q to %q", w, before, after)
		}
	}
}

func TestTheTabsKeepTheirFullNamesForAsLongAsTheyFit(t *testing.T) {
	text := stripText(newLab(t, true, 120, 30))
	for _, name := range []string{"Home", "Projects", "Doctor", "Deps", "Release"} {
		if !strings.Contains(text, name) {
			t.Errorf("at 120 columns %q is not spelled out:\n%s", name, text)
		}
	}
}

func TestTheDoctorTabCarriesTheCountOfFailures(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("3")
	if text := stripText(l); !strings.Contains(text, "Doctor (3)] ×2") {
		t.Fatalf("the Doctor tab should carry its failures:\n%s", text)
	}
	if first := l.lines()[0]; !strings.Contains(first, "×2") {
		t.Fatalf("the header carries the health: %q", first)
	}
}

func TestWarningsAloneNeverShowTheFailureMark(t *testing.T) {
	s := newLab(t, true, 140, 30).sh()
	s.doctor.Report = &doctor.Report{Passed: 3, Warnings: 2}
	for what, got := range map[string]string{"tab": pane.Strip(s.tabBadge("doctor")), "header": pane.Strip(s.healthSummary())} {
		if strings.Contains(got, "×") || !strings.Contains(got, "▲2") {
			t.Errorf("the %s says %q for 0 failed, 2 warnings", what, got)
		}
	}
	s.doctor.Report = &doctor.Report{Failed: 1, Warnings: 2}
	if got := pane.Strip(s.tabBadge("doctor")); got != "×1" {
		t.Errorf("the tab says %q for 1 failed", got)
	}
}

func TestTheTabsAreASingleRowOnAnyHeight(t *testing.T) {
	for _, rows := range []int{23, 24, 40} {
		l := newLab(t, true, 120, rows)
		if got := len(l.sh().tabStrip().Lines); got != 1 {
			t.Fatalf("%d rows: the strip has %d rows", rows, got)
		}
		l.has("4 projects") // the count is said once, in the line under the tabs
	}
}

func TestTheAddProjectWizardIsTitledAddAProject(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("a")
	l.has("Add a project")
	if strings.Contains(l.view(), "─ Add a project ─") || !strings.Contains(l.view(), " Add a project › ") {
		t.Fatalf("the form's title is the first line of the pane, and the rule under the tabs has no title:\n%s", l.view())
	}
}

func TestADryRunNeverShowsAGreenTickOnTheProjectsTab(t *testing.T) {
	s := newLab(t, true, 140, 30).sh()
	now := time.Now()
	s.run = runview.NewRun([]string{"release", "--dry-run"}, nil, now)
	s.run.End(0, false, now)
	got := pane.Strip(s.tabBadge("projects"))
	if strings.Contains(got, "✓") || got == "" {
		t.Errorf("a dry run shows %q", got)
	}
	s.run = runview.NewRun([]string{"run", "build"}, nil, now)
	s.run.End(0, false, now)
	if got := pane.Strip(s.tabBadge("projects")); got != "✓" {
		t.Errorf("a real run that passed shows %q", got)
	}
}

func TestATabShowsASpinnerWhileItsWorkGoesOnInTheBackground(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.depsLoading = true
	l.m = sh
	if tabs := l.lines()[0]; !strings.Contains(tabs, "[Deps (4)] ⠋") {
		t.Fatalf("Deps has no badge while the registries are asked: %q", tabs)
	}
	if !l.sh().animating() {
		t.Fatal("a badge that spins needs ticks")
	}
	sh.depsLoading = false
	l.m = sh
	if strings.Contains(l.lines()[1], "[Deps (4)] ⠋") {
		t.Fatal("the badge stays after the work ended")
	}
}
