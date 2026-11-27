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

func TestTheRunsSectionOpensTheChosenRunInThePanel(t *testing.T) {
	l := newLab(t, true, 130, 36)
	l.next = failingRun()
	l.keys("b")
	l.next = &scripted{lines: []string{"orders:test ▶ go test", "orders:test │ PASS", "orders:test ✓ 12ms"}}
	l.keys("t")
	closeRunPanel(l)
	l.clickText("Runs")
	l.has("✓ test", "passed", "× build", "failed", "just now")
	l.hasNot("Command log")
	// The newest run is chosen; Down moves to the older run and Enter opens it.
	l.keys("down", "enter")
	l.has("run 1 of 2", "$ smith run build")
	// A click on a row only chooses it; a double click opens it.
	closeRunPanel(l)
	x, y := l.find("✓ test")
	l.send(click(x, y))
	if l.sh().projRunAt != 0 || l.sh().runShown() {
		t.Fatalf("the click chose row %d (run panel %v)", l.sh().projRunAt, l.sh().runShown())
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
	l.has("Filter:", "Result")
}

func TestTheRunFooterNamesTheKeysThatMoveBetweenRuns(t *testing.T) {
	l := newLab(t, true, 150, 36)
	l.next = failingRun()
	l.keys("b")
	l.keys("r")
	if f := footerOf(l); !strings.Contains(f, ", older") || strings.Contains(f, "newer") || !strings.Contains(f, "Esc back") {
		t.Fatalf("run footer %q", f)
	}
	closeRunPanel(l)
	l.clickText("Runs")
	if f := footerOf(l); !strings.Contains(f, "Tab section") || !strings.Contains(f, "Enter open") || strings.Contains(f, "follow") {
		t.Fatalf("runs section footer %q", f)
	}
}
