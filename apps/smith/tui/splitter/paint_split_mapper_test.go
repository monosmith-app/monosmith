package splitter

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// styled is a cell as the theme draws it in a look.
func styled(th theme.Theme, look Look, glyph string) string {
	if look == Dragging {
		return th.ButtonActive(glyph)
	}
	return th.Accent(glyph)
}

func TestPaintJunctionsColoursEveryTeeThatTouchesTheDivider(t *testing.T) {
	th := theme.Default()
	col := Spot{ID: "c", Axis: Column, Hit: layout.Rect{X: 4, Y: 1, W: 1, H: 4}}
	for _, tc := range []struct {
		name string
		j    Junction
		p    Spot
	}{
		{"┬ over the column", Junction{X: 4, Y: 0, Glyph: "┬"}, col},
		{"├ on the column", Junction{X: 4, Y: 3, Glyph: "├"}, col},
		{"┤ on the column", Junction{X: 4, Y: 3, Glyph: "┤"}, col},
		{"┼ on the column", Junction{X: 4, Y: 3, Glyph: "┼"}, col},
		{"┴ under the column", Junction{X: 4, Y: 5, Glyph: "┴"}, col},
		{"├ at the left end of a row", Junction{X: 4, Y: 3, Glyph: "├"}, Spot{ID: "r", Axis: Row, Hit: layout.Rect{X: 5, Y: 3, W: 5, H: 1}}},
		{"┤ at the right end of a row", Junction{X: 10, Y: 3, Glyph: "┤"}, Spot{ID: "r", Axis: Row, Hit: layout.Rect{X: 5, Y: 3, W: 5, H: 1}}},
	} {
		for _, look := range []Look{Selected, Dragging} {
			lines := []string{"aaaaaaaaaaaa", "bbbbbbbbbbbb", "cccccccccccc", "dddddddddddd", "eeeeeeeeeeee", "ffffffffffff"}
			lines[tc.j.Y] = strings.Repeat(" ", 12)
			got := PaintJunctions(th, lines, []Junction{tc.j}, tc.p, look)
			if want := styled(th, look, tc.j.Glyph); !strings.Contains(got[tc.j.Y], want) {
				t.Errorf("%s, look %d: the junction is %q, want %q in it", tc.name, look, got[tc.j.Y], want)
			}
		}
		lines := []string{strings.Repeat(" ", 12), strings.Repeat(" ", 12), strings.Repeat(" ", 12), strings.Repeat(" ", 12), strings.Repeat(" ", 12), strings.Repeat(" ", 12)}
		if got := PaintJunctions(th, append([]string(nil), lines...), []Junction{tc.j}, tc.p, Resting); strings.Join(got, "") != strings.Join(lines, "") {
			t.Errorf("%s: a resting divider leaves its junction in the rule's colour", tc.name)
		}
	}
}

func TestPaintJunctionsLeavesTheJunctionsOfOtherDividers(t *testing.T) {
	th := theme.Default()
	col := Spot{ID: "c", Axis: Column, Hit: layout.Rect{X: 4, Y: 1, W: 1, H: 4}}
	far := Junction{X: 9, Y: 0, Glyph: "┬"}
	lines := []string{strings.Repeat(" ", 12)}
	if got := PaintJunctions(th, lines, []Junction{far}, col, Dragging); strings.TrimSpace(got[0]) != "" {
		t.Fatalf("a tee of another divider is painted: %q", got[0])
	}
}
