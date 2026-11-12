package cachewizard

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
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

func TestAShortScreenIsCentredInTheFreeRows(t *testing.T) {
	m := listOf("one", "two").Within(3, 2).Place(100, 40)
	r := m.Rect()
	if r.H >= 22 {
		t.Fatalf("the test needs a screen shorter than the room: %+v", r)
	}
	if got := len(m.Lines(theme.Default(), 0, time.Now())); got != r.H {
		t.Fatalf("the wizard draws %d rows in a %d-row box", got, r.H)
	}
	above, below := r.Y-3, 40-2-(r.Y+r.H)
	if above-below > 1 || below-above > 1 {
		t.Fatalf("the box must sit in the middle of the free rows: %d above, %d below", above, below)
	}
	x, y := cellOf(t, m, "2) two")
	if _, out := m.Click(theme.Default(), x+1, y); out.Reply == nil || out.Reply.Value != "two" {
		t.Fatalf("a click on the centred box must pick the option: %+v", out)
	}
}

// scrollbarOnTheBorder fails unless the lines draw a scrollbar and every cell
// of its track is the last one inside the right border, as in every pane.
func scrollbarOnTheBorder(t *testing.T, lines []string) {
	t.Helper()
	track := []rune(theme.Default().Icons.Track)[0]
	bars := 0
	for _, l := range lines {
		r := []rune(pane.Strip(l))
		for x, c := range r {
			if c != track {
				continue
			}
			if x != len(r)-2 {
				t.Fatalf("the scrollbar is in column %d, not next to the border in column %d: %q", x, len(r)-2, string(r))
			}
			bars++
		}
	}
	if bars == 0 {
		t.Fatalf("the test needs a body that scrolls:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}

func TestTheCacheWizardScrollbarSitsNextToTheBorder(t *testing.T) {
	for _, size := range [][2]int{{100, 16}, {80, 16}} {
		m := listOf("one", "two", "three", "four", "five", "six", "seven", "eight").Within(3, 2).Place(size[0], size[1])
		scrollbarOnTheBorder(t, m.Lines(theme.Default(), 0, time.Now()))
	}
}
