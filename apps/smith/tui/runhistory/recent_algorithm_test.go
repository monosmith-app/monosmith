package runhistory

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

func TestWindowKeepsTheChosenRowInView(t *testing.T) {
	for _, c := range []struct{ at, total, rows, first, n int }{
		{0, 5, 3, 0, 3},
		{4, 5, 3, 2, 3},
		{1, 2, 10, 0, 2},
	} {
		if first, n := Window(c.at, c.total, c.rows); first != c.first || n != c.n {
			t.Errorf("Window(%d, %d, %d) = %d, %d; want %d, %d", c.at, c.total, c.rows, first, n, c.first, c.n)
		}
	}
}

func TestLineFitsItsWidthAndStylesTheChosenRow(t *testing.T) {
	id := func(s string) string { return s }
	r := Row{Pointer: ">", Command: "smith run test", Result: "passed", Took: "1.2s", Ago: "just now"}
	line := Line(r, 70, id, id)
	if pane.Width(line) != 70 || !strings.Contains(line, "smith run test") || !strings.Contains(line, "just now") {
		t.Fatalf("line %q", line)
	}
	r.Selected = true
	mark := func(s string) string { return "[" + s + "]" }
	if got := Line(r, 70, id, mark); !strings.HasPrefix(got, "[") {
		t.Fatalf("chosen row not styled: %q", got)
	}
}
