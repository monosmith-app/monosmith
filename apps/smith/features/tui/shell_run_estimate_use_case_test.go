package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/rundurations"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runner"
)

// heldBuild starts B (build every project) with a process that stays running
// after kit:build started, with the given history of earlier runs.
func heldBuild(t *testing.T, w, h int, history rundurations.History) (*lab, *scripted) {
	t.Helper()
	l := newLab(t, true, w, h)
	sh := l.sh()
	sh.durations = history
	l.m = sh
	p := &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("B")
	return l, p
}

func TestARunWithAHistorySaysAboutHowLongItStillTakes(t *testing.T) {
	history := rundurations.History{"kit:build": {10000}, "orders:build": {10000}, "web:build": {10000}, "sdk:test": {10000}}
	l, _ := heldBuild(t, 150, 44, history)
	l.has("about 40s left")
	l.advance(30 * time.Second)
	l.has("about 30s left")
}

func TestARunWithATaskThatNeverRanGivesNoEstimate(t *testing.T) {
	l, _ := heldBuild(t, 150, 44, rundurations.History{"kit:build": {10000}})
	l.hasNot("left")
}

func TestAPassedTaskAddsItsTimeToTheHistoryAndItIsKept(t *testing.T) {
	l, p := heldBuild(t, 150, 44, nil)
	var saved []rundurations.History
	sh := l.sh()
	sh.saveDurations = func(h rundurations.History) { saved = append(saved, h) }
	l.m = sh
	l.send(outputMsg{proc: p, lines: []string{
		`{"event":"task.finished","task":"kit:build","status":"passed","durationMs":1500}`,
		`{"event":"task.finished","task":"orders:build","status":"cached"}`,
		`{"event":"task.finished","task":"web:build","status":"failed","durationMs":900,"reason":"exited with code 1"}`,
	}, exit: &runner.Exit{Code: 1}})
	if len(saved) != 1 {
		t.Fatalf("the history is kept once the run ends: %d", len(saved))
	}
	if got := saved[0]["kit:build"]; len(got) != 1 || got[0] != 1500 {
		t.Fatalf("kit:build passed in 1.5s: %v", saved[0])
	}
	for _, id := range []string{"orders:build", "web:build"} {
		if _, ok := saved[0][id]; ok {
			t.Fatalf("%s says nothing about how long it takes: %v", id, saved[0])
		}
	}
}

func TestTheEstimateFitsEveryScreen(t *testing.T) {
	history := rundurations.History{"kit:build": {10000}, "orders:build": {10000}, "web:build": {10000}, "sdk:test": {10000}}
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		l, _ := heldBuild(t, size[0], size[1], history)
		lines := l.lines()
		if len(lines) != size[1] {
			t.Fatalf("%v: %d rows", size, len(lines))
		}
		for y, line := range lines {
			if pane.Width(line) != size[0] {
				t.Fatalf("%v: row %d is %d cells: %q", size, y, pane.Width(line), line)
			}
		}
		if !strings.Contains(l.view(), "about 40s left") {
			t.Fatalf("%v: no estimate:\n%s", size, l.view())
		}
	}
}
