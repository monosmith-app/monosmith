package resultdialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestOnASmallScreenTheDialogStaysBetweenTheTabsAndTheFooterAndScrolls(t *testing.T) {
	th := theme.Default()
	m := NewCommitChoice(false).Within(3, 2).Place(100, 16)
	r := m.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the dialog covers the tabs or the footer: %+v", r)
	}
	if got := len(m.Lines(th)); got != r.H {
		t.Fatalf("the dialog draws %d rows in a %d-row box", got, r.H)
	}
	if rowOf(m, "Ask me each time") >= 0 {
		t.Fatal("the test needs a body taller than the dialog")
	}
	m, _ = m.Key("down")
	m, _ = m.Key("down")
	y := rowOf(m, "Ask me each time")
	if y < 0 {
		t.Fatal("the body must follow the cursor to the last choice")
	}
	if !strings.Contains(pane.Strip(m.Lines(th)[r.H-2]), "Esc") {
		t.Error("the hint must stay on the last row")
	}
	m, out := m.Click(th, r.X+10, y)
	if m.Cursor() != 2 || out.CommitChoice != CommitAsk {
		t.Fatalf("a click on the scrolled choice must pick it: cursor %d, %+v", m.Cursor(), out)
	}
}

func TestTheScrollbarNeverTakesTheLastLetterOfAWrappedLine(t *testing.T) {
	for _, m := range []Model{NewCommitChoice(false), NewUndoOff(), NewAfterChange([]string{"deps", "up", "x"}, nil)} {
		m = m.Within(3, 2).Place(100, 16)
		body, _ := m.body(theme.Default())
		for _, l := range body {
			// The inside is W-2 cells: a blank, the text, the scrollbar, a blank.
			if w := pane.Width(l); w > m.Rect().W-5 {
				t.Fatalf("%q is %d cells, so the scrollbar would take its last letter", pane.Strip(l), w)
			}
		}
	}
}

func TestTheWheelScrollsTheResultWithoutMovingTheCursor(t *testing.T) {
	m := NewCommitChoice(false).Within(3, 2).Place(100, 16)
	m = m.Wheel(50)
	if m.Cursor() != 0 || rowOf(m, "Ask me each time") < 0 {
		t.Fatalf("the wheel must show the end and keep the cursor: cursor %d", m.Cursor())
	}
}

// scrollbarOnTheBorder fails unless the lines draw a scrollbar and every cell
// of its track is the last one inside the right border, as in every pane.
func scrollbarOnTheBorder(t *testing.T, lines []string) {
	t.Helper()
	track := []rune(theme.Default().Icons.Track)[0]
	bars := 0
	for _, l := range lines {
		r := []rune(pane.Strip(l))
		for x, c := range r {
			if c != track {
				continue
			}
			if x != len(r)-2 {
				t.Fatalf("the scrollbar is in column %d, not next to the border in column %d: %q", x, len(r)-2, string(r))
			}
			bars++
		}
	}
	if bars == 0 {
		t.Fatalf("the test needs a body that scrolls:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}

func TestTheResultScrollbarSitsNextToTheBorder(t *testing.T) {
	m := NewCommitChoice(false).Within(3, 2).Place(100, 16)
	scrollbarOnTheBorder(t, m.Lines(theme.Default()))
}
