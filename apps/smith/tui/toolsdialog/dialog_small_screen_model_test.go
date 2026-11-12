package toolsdialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// rowOf is the screen row where the dialog draws text, or -1.
func rowOf(m Model, text string) int {
	for i, l := range m.Lines(theme.Default()) {
		if strings.Contains(pane.Strip(l), text) {
			return m.Rect().Y + i
		}
	}
	return -1
}

func TestOnASmallScreenTheToolsDialogScrollsAndKeepsItsKeys(t *testing.T) {
	m := New([]toolinstall.ToolPlan{plan("go", install("winget", false, "winget", "install", "go"), install("scoop", false, "scoop", "install", "go"), install("choco", true, "choco", "install", "go"))}).Within(3, 2).Place(100, 16)
	r := m.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the dialog covers the tabs or the footer: %+v", r)
	}
	if got := len(m.Lines(theme.Default())); got != r.H {
		t.Fatalf("the dialog draws %d rows in a %d-row box", got, r.H)
	}
	m, _ = press(m, "down", "down", "down", "down", "down")
	y := rowOf(m, "Skip it")
	if y < 0 {
		t.Fatal("the body must follow the cursor to the last option")
	}
	if rowOf(m, "Enter") != r.Y+r.H-2 {
		t.Error("the keys must stay on the last row")
	}
	m, _ = m.Click(theme.Default(), r.X+10, y)
	if res := m.Results(); len(res) != 1 || !strings.Contains(res[0].Line, "GO") {
		t.Fatalf("a click on the scrolled option must pick it: %+v", res)
	}
}
