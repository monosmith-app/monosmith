package selection

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

// screen is a 14 by 6 screen with a one-cell frame, as a panel draws it.
var screen = []string{
	"+------------+",
	"|alpha beta  |",
	"|gamma delta |",
	"|            |",
	"|eps zeta    |",
	"+------------+",
}

var inner = layout.Rect{X: 1, Y: 1, W: 12, H: 4}

func TestTextOfOneRowIsTheCellsBetweenTheTwoColumns(t *testing.T) {
	s := State{}.Press(inner, 1, 1).Drag(5, 1)
	if got := s.Text(screen); got != "alpha" {
		t.Fatalf("%q", got)
	}
}

func TestTextOfManyRowsRunsFromTheAnchorToTheHead(t *testing.T) {
	s := State{}.Press(inner, 7, 1).Drag(5, 4)
	if got, want := s.Text(screen), "beta\ngamma delta\n\neps z"; got != want {
		t.Fatalf("%q, want %q", got, want)
	}
}

func TestTextDoesNotDependOnTheDirectionOfTheDrag(t *testing.T) {
	down := State{}.Press(inner, 7, 1).Drag(5, 2)
	up := State{}.Press(inner, 5, 2).Drag(7, 1)
	if down.Text(screen) != up.Text(screen) || down.Text(screen) != "beta\ngamma" {
		t.Fatalf("%q %q", down.Text(screen), up.Text(screen))
	}
}

func TestTextNeverTakesTheFrameCells(t *testing.T) {
	s := State{}.Press(inner, 1, 1).Drag(40, 40)
	for _, c := range "|+-" {
		for _, r := range s.Text(screen) {
			if r == c {
				t.Fatalf("frame cell %q in %q", c, s.Text(screen))
			}
		}
	}
}

func TestTextDropsTheEmptyRowsAtTheEndsAndTrimsTrailingSpaces(t *testing.T) {
	s := State{}.Press(inner, 1, 3).Drag(12, 4)
	if got := s.Text(screen); got != "eps zeta" {
		t.Fatalf("%q", got)
	}
	s = State{}.Press(inner, 1, 2).Drag(12, 2)
	if got := s.Text(screen); got != "gamma delta" {
		t.Fatalf("%q", got)
	}
}

func TestTextStripsStylesAndTakesWideCharactersWhole(t *testing.T) {
	rows := []string{"a\x1b[31m世界\x1b[0mbc", "x"}
	r := layout.Rect{X: 0, Y: 0, W: 7, H: 1}
	// Cells: a=0, 世=1-2, 界=3-4, b=5, c=6.
	if got := (State{}.Press(r, 0, 0).Drag(1, 0)).Text(rows); got != "a世" {
		t.Fatalf("a head on the first half takes the character: %q", got)
	}
	if got := (State{}.Press(r, 2, 0).Drag(3, 0)).Text(rows); got != "世界" {
		t.Fatalf("an anchor on the second half takes the character: %q", got)
	}
	if got := (State{}.Press(r, 5, 0).Drag(6, 0)).Text(rows); got != "bc" {
		t.Fatalf("%q", got)
	}
}

func TestTextIsEmptyWhenNothingIsSelected(t *testing.T) {
	if got := (State{}.Press(inner, 1, 1)).Text(screen); got != "" {
		t.Fatalf("%q", got)
	}
}
