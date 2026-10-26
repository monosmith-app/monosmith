package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

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
		l.next = &scripted{lines: []string{"kit:build ✗ exited with code 2"}, exit: runner.Exit{Code: 1}}
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
	for _, name := range []string{"Home", "Projects", "Run", "Doctor", "Deps", "Release"} {
		if !strings.Contains(text, name) {
			t.Errorf("at 120 columns %q is not spelled out:\n%s", name, text)
		}
	}
}

func TestTheDoctorTabCarriesTheCountOfFailures(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("4")
	if text := stripText(l); !strings.Contains(text, "Doctor ✗2") {
		t.Fatalf("the Doctor tab should carry its failures:\n%s", text)
	}
	if first := l.lines()[0]; !strings.Contains(first, "✗2") {
		t.Fatalf("the header carries the health: %q", first)
	}
}

func TestWarningsAloneNeverShowTheFailureMark(t *testing.T) {
	s := newLab(t, true, 140, 30).sh()
	s.doctor.Report = &doctor.Report{Passed: 3, Warnings: 2}
	for what, got := range map[string]string{"tab": pane.Strip(s.tabBadge("doctor")), "header": pane.Strip(s.healthSummary())} {
		if strings.Contains(got, "✗") || !strings.Contains(got, "▲2") {
			t.Errorf("the %s says %q for 0 failed, 2 warnings", what, got)
		}
	}
	s.doctor.Report = &doctor.Report{Failed: 1, Warnings: 2}
	if got := pane.Strip(s.tabBadge("doctor")); got != "✗1" {
		t.Errorf("the tab says %q for 1 failed", got)
	}
}

func TestOnAShortTerminalTheTabsAreASingleRow(t *testing.T) {
	l := newLab(t, true, 120, 23)
	if got := len(l.sh().tabStrip().Lines); got != 1 {
		t.Fatalf("the strip has %d rows", got)
	}
	l.has("Projects (4)") // the title is on the pane's own edge again
	if l = newLab(t, true, 120, 24); len(l.sh().tabStrip().Lines) != 3 {
		t.Fatal("24 rows is tall enough for boxes")
	}
}

func TestTheAddProjectWizardIsTitledAddAProject(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("a")
	l.has("Add a project")
	if !strings.Contains(stripText(l), "Add a project") {
		t.Fatalf("the form's title belongs in the tab strip:\n%s", l.view())
	}
}
