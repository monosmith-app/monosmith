package tui

import (
	"context"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// slowPlan makes the lab's plan reader wait until it is stopped or released,
// and reports when it was stopped.
func (l *lab) slowPlan() (release chan struct{}, stopped chan struct{}) {
	release, stopped = make(chan struct{}), make(chan struct{})
	sh := l.sh()
	sh.sources.Capture = func(ctx context.Context, argv []string) sources.Captured {
		select {
		case <-ctx.Done():
			close(stopped)
			return sources.Captured{Text: "interrupted"}
		case <-release:
			return sources.Captured{OK: true, Text: "change  smith.toml (defaults vintage)"}
		}
	}
	l.m = sh
	return release, stopped
}

// The dialog opens at once, says what happens, and cannot run before the plan.
func TestTheConfirmationOpensAtOnceAndRunWaitsForThePlan(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	release, _ := l.slowPlan()
	l.ask("upgrade")
	l.has("Upgrade the workspace", "Working out what will change", "smith upgrade --yes")
	l.keys("enter", "y", " ")
	l.has("Working out what will change")
	if len(l.started) != 0 {
		t.Fatalf("Run must wait for the plan: %v", l.started)
	}
	close(release)
}

// Esc stops the work, and the plan that is about to arrive opens nothing.
func TestEscStopsTheWaitForThePlan(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	_, stopped := l.slowPlan()
	l.ask("upgrade")
	l.keys("esc")
	l.hasNot("Working out what will change", "The command that will run:")
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("the work was not stopped")
	}
	l.send(capturedMsg{purpose: capPlan, argv: []string{"upgrade", "--dry-run"}, res: sources.Captured{OK: true, Text: "late plan"}, ticket: 1})
	l.hasNot("late plan", "The command that will run:")
}

// The plan replaces the working dialog and Run then works.
func TestThePlanFillsTheWaitingDialog(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	l.slowPlan()
	l.ask("upgrade")
	l.send(capturedMsg{purpose: capPlan, argv: []string{"upgrade", "--dry-run"}, res: sources.Captured{OK: true, Text: "change  smith.toml (defaults vintage)"}, ticket: 1})
	l.has("change smith.toml")
	l.hasNot("Working out what will change")
	l.keys("enter")
	if got := l.lastStarted(); got != "upgrade --yes" {
		t.Fatalf("started %q", got)
	}
}
