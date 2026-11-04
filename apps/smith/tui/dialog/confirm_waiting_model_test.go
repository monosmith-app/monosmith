package dialog

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// A dialog that waits for its plan shows the message and the command, counts
// the seconds after two, and refuses Run in every way.
func TestAWaitingConfirmShowsProgressAndRefusesRun(t *testing.T) {
	c := NewWaiting("Update dependencies", "This moves them.", []string{"deps", "up", "zod", "--yes"}, "Working out what will change").Within(2, 1).Place(100, 30)
	if !c.Working() {
		t.Fatal("the dialog must be working")
	}
	text := pane.Strip(strings.Join(c.Pulse(3, 5*time.Second).Lines(theme.Default()), "\n"))
	for _, want := range []string{"Working out what will change", "5s", "$ smith deps up zod --yes", "Cancel"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	for _, key := range []string{"enter", " ", "y", "Y", "tab", "c"} {
		if _, out := c.Key(key); out.Done {
			t.Fatalf("%q must do nothing while the plan is read", key)
		}
	}
	if _, out := c.Click(c.Rect().X+c.buttonsRow(c.Rect().W - 4)[0][0] + 1, c.Rect().Y+c.Rect().H-2); out.Done {
		t.Fatal("a click on Run must do nothing")
	}
	if _, out := c.Key("esc"); !out.Done || out.Confirmed {
		t.Fatalf("esc must cancel: %+v", out)
	}
}

// A dialog that only reads says the command is running, not that it will run.
func TestAReadingConfirmSaysTheCommandIsRunning(t *testing.T) {
	c := NewWaiting("Add a project", "smith asks the plugins.", []string{"choices", "--json"}, "Reading the choices").Reading().Within(2, 1).Place(100, 30)
	text := pane.Strip(strings.Join(c.Lines(theme.Default()), "\n"))
	if !strings.Contains(text, "The command that is running:") || strings.Contains(text, "will run") {
		t.Fatalf("the caption is wrong:\n%s", text)
	}
}
