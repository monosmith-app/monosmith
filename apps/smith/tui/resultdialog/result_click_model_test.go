package resultdialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// rowOf is the screen row where the dialog draws text.
func rowOf(m Model, text string) int {
	for i, l := range m.Lines(theme.Default()) {
		if strings.Contains(pane.Strip(l), text) {
			return m.Rect().Y + i
		}
	}
	return -1
}

func TestAClickPicksTheChoiceUnderThePointerAndADoubleClickSelectsIt(t *testing.T) {
	m := NewAfterChange([]string{"deps", "up", "x"}, nil).Place(120, 36)
	th := theme.Default()
	y := rowOf(m, "Commit only the files")
	if y < 0 {
		t.Fatal("choice not drawn")
	}
	m, out := m.Click(th, m.Rect().X+10, y, false)
	if m.Cursor() != 1 || out.Commit != nil || out.Done {
		t.Fatalf("a single click must only move the cursor: cursor=%d out=%+v", m.Cursor(), out)
	}
	if _, out = m.Click(th, m.Rect().X+10, y, true); out.Commit == nil || out.Commit[len(out.Commit)-1] != "--only-operation" {
		t.Fatalf("a double click must select the choice: %+v", out)
	}
}

func TestAClickOutsideTheChoicesChangesNothing(t *testing.T) {
	m := NewAfterChange([]string{"deps", "up", "x"}, nil).Place(120, 36)
	th := theme.Default()
	for _, p := range [][2]int{{0, 0}, {m.Rect().X + 5, m.Rect().Y}, {m.Rect().X + 5, rowOf(m, "Nothing is committed")}} {
		got, out := m.Click(th, p[0], p[1], true)
		if got.Cursor() != 0 || out.Done || out.Commit != nil {
			t.Fatalf("click at %v changed the dialog: cursor=%d out=%+v", p, got.Cursor(), out)
		}
	}
}

func TestTheDialogShowsTheMouseHint(t *testing.T) {
	m := NewAfterChange([]string{"deps", "up", "x"}, nil).Place(120, 36)
	if rowOf(m, "click an entry, double-click to select") < 0 {
		t.Fatal("hint missing")
	}
}
