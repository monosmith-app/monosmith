package selection

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

func TestDivideCutsAnAreaAtAColumnAndAtARowOfOneSide(t *testing.T) {
	area := layout.Rect{X: 1, Y: 1, W: 20, H: 10}
	// Settings: a column at x=8, then a row at y=6 that spans the right side only.
	got := Divide([]layout.Rect{area}, []Cut{{X: 9, Y: 6}, {X: 8, Y: 1, Vertical: true}})
	want := []layout.Rect{
		{X: 1, Y: 1, W: 7, H: 10},
		{X: 9, Y: 1, W: 12, H: 5},
		{X: 9, Y: 7, W: 12, H: 4},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("%v, want %v", got, want)
	}
}

func TestDivideLeavesAnUntouchedAreaWholeAndDropsEmptyHalves(t *testing.T) {
	a, b := layout.Rect{X: 0, Y: 0, W: 5, H: 5}, layout.Rect{X: 10, Y: 0, W: 5, H: 5}
	got := Divide([]layout.Rect{a, b}, []Cut{{X: 0, Y: 0, Vertical: true}, {X: 99, Y: 99}})
	want := []layout.Rect{{X: 1, Y: 0, W: 4, H: 5}, b}
	if !slices.Equal(got, want) {
		t.Fatalf("%v, want %v", got, want)
	}
}
