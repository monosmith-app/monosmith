package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// inColour draws escape codes for the length of a test, so the looks can be told apart.
func inColour(t *testing.T) {
	was := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(was) })
}

func TestARowUnderThePointerTakesTheRowHoverLookAndLosesItWhenThePointerLeaves(t *testing.T) {
	inColour(t)
	l := buttonLab(t, 160, 40)
	x, y := l.find("○ kit")
	rest := l.rawRow(y)
	l.send(motion(x+1, y))
	if l.rawRow(y) == rest {
		t.Fatal("the row under the pointer must show the hover look")
	}
	if want := l.sh().th.RowHover(""); want == "" {
		t.Fatal("the colour theme has a row hover look")
	}
	l.send(motion(0, y))
	if got := l.rawRow(y); got != rest {
		t.Fatalf("the row must be back at rest:\n%q\n%q", got, rest)
	}
}

func TestTheSelectedRowKeepsItsOwnLookUnderThePointer(t *testing.T) {
	inColour(t)
	l := buttonLab(t, 160, 40)
	x, y := l.find("● web")
	rest := l.rawRow(y)
	l.send(motion(x+1, y))
	if got := l.rawRow(y); got != rest {
		t.Fatalf("the selected row must not change under the pointer:\n%q\n%q", got, rest)
	}
}

func TestNoRowBehindADialogLooksHovered(t *testing.T) {
	l := buttonLab(t, 160, 40)
	x, y := l.find("○ kit")
	l.keys("?") // the shortcuts overlay is a modal
	if !l.sh().modalOpen() {
		t.Skip("the help overlay is not a modal in this lab")
	}
	sh := l.sh()
	sh.mouseX, sh.mouseY = x+1, y
	in := sh.mainInner()
	c := sh.page().View(in.W, in.H)
	if got := sh.withRowHover(c, in); got.lines[y-in.Y] != c.lines[y-in.Y] {
		t.Fatal("a row behind a dialog must not take the hover look")
	}
}
