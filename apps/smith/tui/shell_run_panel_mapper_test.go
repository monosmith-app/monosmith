package tui

import (
	"strings"
	"testing"
)

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

func TestTheChosenTaskHasTheCursorMarkOfEveryPickListAndOnlyIt(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	l.clickText("web:build")
	marked := 0
	for _, line := range l.lines() {
		// The run panel is the right pane.
		parts := strings.Split(line, "│")
		line = parts[len(parts)-1]
		if strings.Contains(line, "› ") && (strings.Contains(line, ":build") || strings.Contains(line, "All output")) {
			marked++
			if !strings.Contains(line, "web:build") {
				t.Fatalf("the mark is on another row: %q", line)
			}
		}
	}
	if marked != 1 {
		t.Fatalf("%d task rows carry the cursor mark, want 1", marked)
	}
}
