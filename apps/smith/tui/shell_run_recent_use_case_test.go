package tui

import (
	"strings"
	"testing"
	"time"
)

func TestTheStartScreenListsTheKeptRunsAndOpensTheChosenOne(t *testing.T) {
	l := newLab(t, true, 100, 36)
	l.next = failingRun()
	l.keys("b")
	l.next = &scripted{lines: []string{"kit:test ▶ go test", "kit:test │ PASS", "kit:test ✓ 12ms"}}
	l.keys("t")
	l.keys("esc", "esc")
	l.has("Recent runs", "smith run test", "passed", "smith run build", "failed", "just now")
	l.hasNot("Command log")
	// The newest run is chosen; Down moves to the older one and Enter opens it.
	l.keys("down", "enter")
	l.has("run 1 of 2", "$ smith run build")
	// A click on a row only chooses it; a double click opens it.
	l.keys("esc", "esc")
	x, y := l.find("✓ passed")
	l.send(click(x, y))
	l.has("Recent runs")
	if l.sh().recentAt != 0 {
		t.Fatalf("the click chose row %d", l.sh().recentAt)
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
	l.has("‹ Start · run 2 of 2 · ‹ older  newer ›")
	l.clickText("‹ older")
	l.has("run 1 of 2", "$ smith run build")
	// At the oldest run the older part is disabled and says why.
	l.clickText("‹ older")
	l.has("This is the oldest run kept")
	l.clickText("newer ›")
	l.has("run 2 of 2")
	l.clickText("‹ Start")
	l.has("Recent runs", "Verify everything")
}

func TestTheRunFooterNamesTheKeysThatMoveBetweenRuns(t *testing.T) {
	l := newLab(t, true, 150, 36)
	l.next = failingRun()
	l.keys("b")
	l.keys("r")
	if f := footerOf(l); !strings.Contains(f, "older (,)") || strings.Contains(f, "newer") || !strings.Contains(f, "start screen (Esc)") {
		t.Fatalf("run footer %q", f)
	}
	l.keys("esc", "esc")
	if f := footerOf(l); !strings.Contains(f, "↑↓ choose run") || !strings.Contains(f, "open (Enter)") || strings.Contains(f, "follow") {
		t.Fatalf("start screen footer %q", f)
	}
}
