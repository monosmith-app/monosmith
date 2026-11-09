package tui

import (
	"strings"
	"testing"
	"time"
)

// closeRunPanel presses Esc until the right panel shows the project again.
func closeRunPanel(l *lab) {
	for i := 0; i < 3 && l.sh().runShown(); i++ {
		l.keys("esc")
	}
}

func TestTheRecentRunsUnderTheListOpenTheChosenOneInThePanel(t *testing.T) {
	l := newLab(t, true, 100, 36)
	l.next = failingRun()
	l.keys("b")
	l.next = &scripted{lines: []string{"kit:test ▶ go test", "kit:test │ PASS", "kit:test ✓ 12ms"}}
	l.keys("t")
	closeRunPanel(l)
	l.has("Recent runs", "smith run test", "passed", "smith run build", "failed", "just now")
	l.hasNot("Command log")
	// The newest run is chosen; Tab moves to the list, Down to the older run and Enter opens it.
	l.keys("tab", "down", "enter")
	l.has("run 1 of 2", "$ smith run build")
	// A click on a row only chooses it; a double click opens it.
	closeRunPanel(l)
	x, y := l.find("✓ passed")
	l.send(click(x, y))
	l.has("Recent runs")
	if l.sh().recentAt != 0 || l.sh().runShown() {
		t.Fatalf("the click chose row %d (run panel %v)", l.sh().recentAt, l.sh().runShown())
	}
	l.advance(time.Second)
	l.doubleClick(x, y)
	l.has("run 2 of 2", "$ smith run test")
}

func TestTheRunHeaderStepsThroughTheRunsWithTheMouse(t *testing.T) {
	l := newLab(t, true, 100, 36)
	l.next = failingRun()
	l.keys("b")
	l.next = &scripted{lines: []string{"kit:test ✓ 12ms"}}
	l.keys("t")
	l.has("run 2 of 2 · ‹ older  newer ›")
	l.clickText("‹ older")
	l.has("run 1 of 2", "$ smith run build")
	// At the oldest run the older part is disabled and says why.
	l.clickText("‹ older")
	l.has("This is the oldest run kept")
	l.clickText("newer ›")
	l.has("run 2 of 2")
	l.clickText("Back (Esc)")
	l.has("Recent runs", "Verify everything")
}

func TestTheRunFooterNamesTheKeysThatMoveBetweenRuns(t *testing.T) {
	l := newLab(t, true, 150, 36)
	l.next = failingRun()
	l.keys("b")
	l.keys("r")
	if f := footerOf(l); !strings.Contains(f, "older (,)") || strings.Contains(f, "newer") || !strings.Contains(f, "back (Esc)") {
		t.Fatalf("run footer %q", f)
	}
	closeRunPanel(l)
	l.keys("tab")
	if f := footerOf(l); !strings.Contains(f, "↑↓ choose run") || !strings.Contains(f, "open (Enter)") || strings.Contains(f, "follow") {
		t.Fatalf("recent runs footer %q", f)
	}
}
