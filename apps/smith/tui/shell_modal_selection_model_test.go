package tui

import (
	"strings"
	"testing"
)

func TestDraggingOverTheTextOfADialogCopiesItAndKeepsTheHighlight(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("?")
	out := copying(l)
	r, ok := l.sh().modalRect()
	if !ok {
		t.Fatal("the help dialog is open")
	}
	// Any row inside the dialog that holds letters.
	var x, y int
	for row, line := range l.lines() {
		if row > r.Y+1 && row < r.Y+r.H-1 && strings.TrimSpace(line[min(r.X+2, len(line)):]) != "" {
			y, x = row, r.X+2
			break
		}
	}
	if y == 0 {
		t.Fatalf("no text row in the dialog:\n%s", l.view())
	}
	l.send(click(x, y)).send(drag(x+4, y)).send(drag(x+8, y))
	if !l.sh().selection.Active {
		t.Fatal("a drag in the dialog selects")
	}
	l.send(letGo(x+8, y))
	if got := l.copied(t, out); got == "" {
		t.Fatal("the release copies the selected text")
	}
	if !l.sh().selection.Held {
		t.Fatal("the highlight stays after the release")
	}
	l.send(drag(x+9, y)) // a pointer move keeps the highlight
	if !l.sh().selection.Held {
		t.Fatal("moving the pointer must not drop the highlight")
	}
}

func TestAClickInADialogWithoutMovementSelectsNothing(t *testing.T) {
	l := newLab(t, true, 140, 40)
	l.keys("?")
	out := copying(l)
	r, _ := l.sh().modalRect()
	l.send(click(r.X+3, r.Y+3)).send(letGo(r.X+3, r.Y+3))
	if out.Len() != 0 || l.sh().selection.Selected() {
		t.Fatal("a plain click in a dialog selects nothing")
	}
}
