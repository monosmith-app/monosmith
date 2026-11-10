package selection

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

var region = layout.Rect{X: 2, Y: 1, W: 10, H: 4}

func TestAClickWithoutMovementSelectsNothing(t *testing.T) {
	s := State{}.Press(region, 4, 2)
	if s.Selected() {
		t.Fatal("a press alone selects nothing")
	}
	s = s.Drag(4, 2)
	got, ok := s.Release()
	if ok || got != (State{}) {
		t.Fatalf("a click leaves no selection: %+v %v", got, ok)
	}
}

func TestDraggingOneCellSelectsAndTheReleaseKeepsTheHighlight(t *testing.T) {
	s := State{}.Press(region, 4, 2).Drag(5, 2)
	if !s.Selected() || !s.Active {
		t.Fatalf("a drag of one cell selects: %+v", s)
	}
	got, ok := s.Release()
	if !ok || !got.Held || got.Active || !got.Selected() {
		t.Fatalf("the release keeps the highlight: %+v %v", got, ok)
	}
	if got.Clear().Selected() {
		t.Fatal("Clear drops the highlight")
	}
}

func TestTheHeadStaysInsideTheRegion(t *testing.T) {
	s := State{}.Press(region, 4, 2).Drag(99, 99)
	if s.Head != (Cell{X: 11, Y: 4}) {
		t.Fatalf("head clamps to the corner: %+v", s.Head)
	}
	s = s.Drag(-5, -5)
	if s.Head != (Cell{X: 2, Y: 1}) {
		t.Fatalf("head clamps to the other corner: %+v", s.Head)
	}
}

func TestDragDoesNothingWithoutAPress(t *testing.T) {
	if s := (State{}).Drag(3, 3); s.Selected() || s != (State{}) {
		t.Fatalf("no press, no drag: %+v", s)
	}
}

func TestReadingOrderIsTheSameWhicheverWayTheDragGoes(t *testing.T) {
	a := State{}.Press(region, 3, 1).Drag(5, 3)
	b := State{}.Press(region, 5, 3).Drag(3, 1)
	fa, la := a.bounds()
	fb, lb := b.bounds()
	if fa != fb || la != lb {
		t.Fatalf("a %v %v, b %v %v", fa, la, fb, lb)
	}
}
