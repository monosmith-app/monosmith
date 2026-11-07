package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

// A run the person stopped is neither passed nor failed: the bar counts what
// ended and what was left, no task is red, and the status line stops saying
// "Stopping" as soon as the process has ended.
func TestAStoppedRunIsCancelledAndNotFailed(t *testing.T) {
	l := newLab(t, true, 140, 34)
	p := &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("B", "ctrl+c")
	l.has("Stopping")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.hasNot("Stopping", "100%")
	l.hasNot("× ")
	l.has("done,", "skipped", "cancelled")
	c := l.sh().run.Counts()
	if c.Failed != 0 || c.Cancelled != 1 {
		t.Fatalf("a stopped task is cancelled, not failed: %+v", c)
	}
	if strings.Contains(l.sh().statusInfo().Hint, "top") {
		t.Fatalf("the status line still says it is stopping: %q", l.sh().statusInfo().Hint)
	}
}

// A run stopped before any task was known does not say it failed.
func TestAStoppedRunWithNoTasksSaysItWasStopped(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.sources.Plan = nil // no task is known up front
	l.m = sh
	p := &scripted{lines: []string{"building"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("w", "ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("stopped after")
	l.hasNot("failed (code")
}
