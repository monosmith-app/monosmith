package pane

import "testing"

func TestBarAtFindsTheColumnOfTrackAndThumbAndNothingElse(t *testing.T) {
	screen := []string{"a ", "a█", "a█", "a┆", "a┆", "a "}
	cell := func(x, y int) string {
		if y < 0 || y >= len(screen) {
			return ""
		}
		r := []rune(screen[y])
		if x < 0 || x >= len(r) {
			return ""
		}
		return string(r[x])
	}
	b, ok := BarAt(cell, 1, 4, "┆", "█")
	if !ok || b != (Bar{Top: 1, H: 4, At: 0, Size: 2}) {
		t.Fatalf("got %+v %v", b, ok)
	}
	if _, ok := BarAt(cell, 0, 2, "┆", "█"); ok {
		t.Fatal("a cell off the bar is no bar")
	}
	screen = []string{"██", "  "}
	if _, ok := BarAt(cell, 0, 0, "┆", "█"); ok {
		t.Fatal("a thumb with no track (a progress bar) is no scrollbar")
	}
}
