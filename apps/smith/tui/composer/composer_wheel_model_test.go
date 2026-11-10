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
