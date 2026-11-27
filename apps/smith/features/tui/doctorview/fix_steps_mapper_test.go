package doctorview

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

func TestTheStepsLightTheCurrentOne(t *testing.T) {
	th := theme.Default()
	got := Steps(th, StepApply)
	if pane.Strip(got) != "1 Review ── 2 Apply ── 3 Check again" {
		t.Fatalf("%q", pane.Strip(got))
	}
	if !strings.Contains(got, th.Bold(th.Accent("2 Apply"))) {
		t.Fatalf("the current step is not lit: %q", got)
	}
}

func TestTheStateLineNamesTheStateWithItsMark(t *testing.T) {
	f := finding("lock file", "")
	f.Warning, f.Project = true, "web"
	if got := pane.Strip(StateLine(theme.Default(), f)); got != "▲ Warning · web" {
		t.Fatalf("%q", got)
	}
}

func TestGroupByStateFilesTheFailuresFirstThenTheWarningsThenThePasses(t *testing.T) {
	warn := finding("lock file", "")
	warn.Warning = true
	ok := finding("git", "")
	ok.OK = true
	got := GroupByState([]doctor.Finding{warn, finding("go.work", ""), ok, finding("hooks", "")})
	want := []Group{{State: StateFailed, Indexes: []int{1, 3}}, {State: StateWarning, Indexes: []int{0}}, {State: StatePassed, Indexes: []int{2}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestByStateTheListHasStateHeadingsTheFixTagAndTheProjectOnASecondLine(t *testing.T) {
	m := sample()
	m.ByState = true
	m.Cursor = m.firstFinding()
	lines, rowOf := m.View(plainTheme, 70, false)
	text := strings.Join(lines, "\n")
	for _, want := range []string{"Failed  3", "Docker registries of web agree", "web · apps/web/Dockerfile", "fix"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	// The second line of a check is the check: a click on it selects it.
	for i, l := range lines {
		if strings.Contains(l, "web · apps/web/Dockerfile") && rowOf[i] != rowOf[i-1] {
			t.Fatalf("the second line maps to %d, the first to %d", rowOf[i], rowOf[i-1])
		}
	}
	// Folding a state group keeps its own key.
	m = m.toggleGroup(0)
	if open, set := m.Folds["state:failed"]; !set || open {
		t.Fatalf("folds %v", m.Folds)
	}
}
