package pane

import "testing"

func TestKeepVisibleNeverHidesTheRowUnderAMoreMarker(t *testing.T) {
	const total, h = 30, 8
	for top := 0; top <= total-h; top++ {
		for row := 0; row < total; row++ {
			got := KeepVisible(top, row, total, h)
			if got < 0 || got > total-h {
				t.Fatalf("top %d row %d: window starts at %d", top, row, got)
			}
			first, last := got, got+h-1
			if got > 0 {
				first++
			}
			if got+h < total {
				last--
			}
			if row < first || row > last {
				t.Fatalf("top %d row %d: window %d shows rows %d to %d", top, row, got, first, last)
			}
		}
	}
}

func TestKeepVisibleLeavesAWindowThatAlreadyShowsTheRow(t *testing.T) {
	if got := KeepVisible(5, 8, 30, 8); got != 5 {
		t.Errorf("moved to %d", got)
	}
}

func TestKeepVisibleShowsEverythingWhenItFits(t *testing.T) {
	if got := KeepVisible(3, 4, 6, 8); got != 0 {
		t.Errorf("got %d", got)
	}
}
