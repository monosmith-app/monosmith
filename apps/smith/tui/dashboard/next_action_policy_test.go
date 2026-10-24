package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheNextActionIsTheMostUrgentThingWithItsKey(t *testing.T) {
	cases := []struct {
		name string
		f    Facts
		text string
		cmd  string
	}{
		{"waiting for doctor", Facts{}, "checking the workspace", ""},
		{"failing checks come first", Facts{Checked: true, Failed: 2, Warnings: 2, Affected: 3, Last: &Run{}}, "2 checks fail: press d", "ws.doctor"},
		{"one failing check", Facts{Checked: true, Failed: 1}, "1 check fails: press d", "ws.doctor"},
		{"a failed run", Facts{Checked: true, Last: &Run{Command: "run lint"}, Affected: 1}, "the last run failed: press 3", "view.run"},
		{"a stopped run is not a failure", Facts{Checked: true, Last: &Run{Cancelled: true}, Affected: 3}, "3 projects affected: press w", "run.affected"},
		{"affected projects", Facts{Checked: true, Affected: 3, Last: &Run{OK: true}}, "3 projects affected: press w to lint and test them", "run.affected"},
		{"warnings only", Facts{Checked: true, Warnings: 2}, "2 warnings: press d", "ws.doctor"},
		{"outdated dependencies", Facts{Checked: true, DepsChecked: true, Outdated: 5}, "5 dependencies have newer releases: press 5", "view.deps"},
		{"nothing to do", Facts{Checked: true, DepsChecked: true}, "all good: press b to build", "run.build"},
		{"doctor failed to run", Facts{CheckErr: "no git"}, "doctor could not run", "ws.doctor"},
	}
	for _, c := range cases {
		got := NextAction(c.f)
		if !strings.Contains(got.Text, c.text) || got.CommandID != c.cmd {
			t.Errorf("%s: %+v, want %q and %q", c.name, got, c.text, c.cmd)
		}
	}
}

func TestThePageLeadsWithTheHeadlineThenTheNextActionThenTheFacts(t *testing.T) {
	f := Facts{
		Checked: true, Failed: 2, Warnings: 2, Passed: 14, ProjectsLoaded: true, Projects: 8, Affected: 3,
		Last: &Run{Command: "run lint test --affected", OK: true, Took: "4.2s"},
	}
	lines, actionRow, id := Lines(theme.Default(), f, 80)
	text := make([]string, len(lines))
	for i, l := range lines {
		text[i] = pane.Strip(l)
	}
	joined := strings.Join(text, "\n")
	if !strings.Contains(text[1], "2 checks fail") || !strings.Contains(text[1], "2 warnings") {
		t.Fatalf("headline: %q", text[1])
	}
	if !strings.Contains(text[actionRow], "Next: 2 checks fail: press d") || id != "ws.doctor" {
		t.Fatalf("next action at %d: %q (%s)", actionRow, text[actionRow], id)
	}
	for _, want := range []string{"3 of 8 projects changed", "passed  run lint test --affected  4.2s", "not checked: 5 opens Deps"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
	for _, l := range lines {
		if pane.Width(l) > 80 {
			t.Errorf("line wider than the pane: %q", pane.Strip(l))
		}
	}
}

func TestAHealthyWorkspaceSaysSo(t *testing.T) {
	lines, _, _ := Lines(theme.Default(), Facts{Checked: true, Passed: 9, ProjectsLoaded: true, Projects: 3, DepsChecked: true}, 60)
	if got := pane.Strip(lines[1]); !strings.Contains(got, "healthy") {
		t.Fatalf("%q", got)
	}
}
