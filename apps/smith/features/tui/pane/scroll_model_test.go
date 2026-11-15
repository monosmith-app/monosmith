package pane

import "testing"

func TestTheWheelMovesTheWindowAndNeverTheSelection(t *testing.T) {
	s := &Scroll{}
	if top := s.Show(2, 10, 100); top != 0 {
		t.Fatalf("a selection in view starts at the top, got %d", top)
	}
	s.By(30)
	if top := s.Show(2, 10, 100); top != 30 {
		t.Fatalf("the wheel scrolled to %d, want 30, with the selection out of view", top)
	}
	s.By(500)
	if top := s.Show(2, 10, 100); top != 90 {
		t.Fatalf("the wheel stops at the last page, got %d", top)
	}
	s.By(-1000)
	if top := s.Show(2, 10, 100); top != 0 {
		t.Fatalf("and at the first, got %d", top)
	}
}

func TestAClickOnARowInViewDoesNotMoveTheWindowAndAKeyScrollsJustEnough(t *testing.T) {
	s := &Scroll{}
	s.Show(0, 10, 100)
	s.By(20)
	s.Show(0, 10, 100)
	if top := s.Show(25, 10, 100); top != 20 {
		t.Fatalf("selecting row 25 of the rows 20-29 moved the window to %d", top)
	}
	if top := s.Show(30, 10, 100); top != 21 {
		t.Fatalf("one row below the window scrolls one row, got %d", top)
	}
	if top := s.Show(29, 10, 100); top != 21 {
		t.Fatalf("moving back into view leaves the window, got %d", top)
	}
	if top := s.Show(5, 10, 100); top != 5 {
		t.Fatalf("a selection above the window brings it to the top, got %d", top)
	}
}

func TestAListShorterThanTheWindowNeverScrolls(t *testing.T) {
	var s *Scroll
	if s.Show(3, 10, 4) != 0 {
		t.Fatal("a nil Scroll scrolls")
	}
	s = &Scroll{}
	s.By(5)
	if s.Show(3, 10, 4) != 0 {
		t.Fatal("a short list scrolled")
	}
}
