package tui

import "testing"

func TestTheButtonsOfTheRunPanelKeepTheirRowsWhateverTaskIsChosen(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	before := l.sh().runGeom(l.sh().detailRect().W, l.sh().detailRect().H).ctlRows
	// web:build has no output: Copy log stays, faint, so no row moves.
	l.clickText("web:build")
	if after := l.sh().runGeom(l.sh().detailRect().W, l.sh().detailRect().H).ctlRows; after != before {
		t.Fatalf("the button rows went from %d to %d", before, after)
	}
	l.has("Copy log (Y)")
}
