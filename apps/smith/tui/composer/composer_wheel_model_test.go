package composer

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func framed(m Model) string {
	lines := m.Lines(theme.Default())
	for i := range lines {
		lines[i] = pane.Strip(lines[i])
	}
	return strings.Join(lines, "\n")
}

func TestTheWheelScrollsTheCatalogueAndTheCursorStays(t *testing.T) {
	m := New(catalogue()).Place(80, 14) // a list longer than the window
	before := framed(m)
	if !strings.Contains(before, "↓ ") {
		t.Skipf("the catalogue fits in this window:\n%s", before)
	}
	m = m.Wheel(4)
	after := framed(m)
	if m.cursor != 0 {
		t.Fatalf("the wheel moved the cursor to %d", m.cursor)
	}
	if after == before || !strings.Contains(after, "↑ ") {
		t.Fatalf("the wheel did not scroll the list:\n%s", after)
	}
}

// TestTheScrollbarOfTheListSitsNextToTheBorder: as in every pane, the bar is
// the last cell inside the frame, with no blank cell between it and the border.
func TestTheScrollbarOfTheListSitsNextToTheBorder(t *testing.T) {
	m := New(catalogue()).Place(80, 14)
	th := theme.Default()
	track, thumb := []rune(th.Icons.Track)[0], []rune(th.Icons.Thumb)[0]
	bars := 0
	for _, line := range strings.Split(framed(m), "\n") {
		r := []rune(line)
		if len(r) < 3 {
			continue
		}
		if r[len(r)-2] == track || r[len(r)-2] == thumb {
			bars++
		}
		if (r[len(r)-3] == track || r[len(r)-3] == thumb) && r[len(r)-2] == ' ' {
			t.Fatalf("the scrollbar sits one cell from the border: %q", line)
		}
	}
	if bars == 0 {
		t.Fatalf("no scrollbar next to the border:\n%s", framed(m))
	}
}
