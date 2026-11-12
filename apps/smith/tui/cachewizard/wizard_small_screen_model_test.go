package cachewizard

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestOnASmallScreenTheWizardScrollsItsListAboveTheButtons(t *testing.T) {
	m := listOf("one", "two", "three", "four", "five", "six", "seven", "eight").Within(3, 2).Place(100, 16)
	r := m.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the wizard covers the tabs or the footer: %+v", r)
	}
	if got := len(m.Lines(theme.Default(), 0, time.Now())); got > r.H {
		t.Fatalf("the wizard draws %d rows in a %d-row box", got, r.H)
	}
	cellOf(t, m, "Cancel (Esc)")
	for range 7 {
		m, _ = m.Key(tea.KeyMsg{Type: tea.KeyDown})
	}
	x, y := cellOf(t, m, "8) eight")
	cellOf(t, m, "Cancel (Esc)")
	_, out := m.Click(theme.Default(), x+1, y)
	if out.Reply == nil || out.Reply.Value != "eight" {
		t.Fatalf("a click on the scrolled option must pick it: %+v", out)
	}
}
