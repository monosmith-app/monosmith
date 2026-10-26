package tui

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

func decodedClip(t *testing.T, out *bytes.Buffer) string {
	t.Helper()
	s := strings.TrimSuffix(strings.TrimPrefix(out.String(), "\x1b]52;c;"), "\a")
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("clipboard %q: %v", out.String(), err)
	}
	return string(b)
}

func TestATaskRunNeverHandsTheTerminalOver(t *testing.T) {
	for _, keys := range [][]string{{"b"}, {"t"}, {"l"}, {"c"}, {"w"}} {
		l := newLab(t, true, 140, 30)
		l.keys(keys...)
		if len(l.started) != 1 || l.sh().quit || len(l.sh().Argv) != 0 {
			t.Fatalf("%v: started %v, quit=%v argv=%v", keys, l.started, l.sh().quit, l.sh().Argv)
		}
		if l.sh().tab != "run" {
			t.Fatalf("%v showed %q, not the Run tab", keys, l.sh().tab)
		}
	}
	// The palette's entries and the Home action take the same way.
	l := newLab(t, true, 140, 30)
	l.keys("ctrl+p").typeText("Lint all").keys("enter")
	if l.lastStarted() != "run lint" || l.sh().quit {
		t.Fatalf("palette started %q", l.lastStarted())
	}
}

func TestEveryRunOfTheSessionIsKeptAndCanBeShownAgain(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	l.next = &scripted{lines: []string{"kit:test ▶ go test", "kit:test │ PASS", "kit:test ✔ 12ms"}}
	l.keys("t")
	l.has("$ smith run test", "run 2/2", "kit:test")
	l.keys(",")
	l.has("$ smith run build", "run 1/2", "orders:build")
	// The earlier run opens on its first failure.
	l.has("Log: orders:build", "undefined: Foo")
	l.keys(".")
	l.has("run 2/2", "kit:test")
	l.keys(".")
	l.has("unavailable: this is the latest run")
	if len(l.sh().runs) != 2 {
		t.Fatalf("%d runs kept", len(l.sh().runs))
	}
}

func TestOnlyTheLatestRunsCountIsKeptInMemory(t *testing.T) {
	l := newLab(t, true, 140, 30)
	for i := 0; i < maxRunsKept+5; i++ {
		l.keys("b")
	}
	if n := len(l.sh().runs); n != maxRunsKept {
		t.Fatalf("%d runs kept, want %d", n, maxRunsKept)
	}
	if l.sh().runAt != maxRunsKept-1 {
		t.Fatalf("shown run %d", l.sh().runAt)
	}
}

func TestTheHistoryKeysWaitWhileARunIsGoing(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("b")
	p := &scripted{lines: []string{"kit:build ▶ x"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("t", ",")
	l.has("a run is in progress")
	if l.sh().runAt != 1 {
		t.Fatalf("browsed to run %d during a run", l.sh().runAt)
	}
}

func TestRunTheFailedTasksAgainStartsOnlyThoseTasks(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	l.keys("R")
	// orders:build failed and web:build was skipped because of it.
	if got := l.lastStarted(); got != "run build -p orders,web" {
		t.Fatalf("started %q", got)
	}
	// Nothing failed in a run that passed: the key says so and starts nothing.
	n := len(l.started)
	l.keys("R")
	l.has("unavailable: no task failed or was skipped")
	if len(l.started) != n {
		t.Fatalf("started %v", l.started)
	}
}

func TestACancelledRunCanBeContinuedWithTheFailedKey(t *testing.T) {
	l := newLab(t, true, 140, 36)
	p := &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("b", "ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("cancelled")
	if got := l.sh().failedArgv(); got == nil || got[0] != "run" {
		t.Fatalf("failed argv %v", got)
	}
	l.keys("R")
	if l.lastStarted() != "run build test -p kit,orders,web,sdk" || len(l.started) != 2 {
		t.Fatalf("started %v", l.started)
	}
}

func TestYCopiesTheCommandLineAndShiftYCopiesTheLogShown(t *testing.T) {
	l := newLab(t, true, 140, 36)
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh
	l.next = failingRun()
	l.keys("b")
	l.keys("y")
	if got := decodedClip(t, &out); got != "smith run build -p orders" {
		t.Fatalf("y copied %q", got)
	}
	out.Reset()
	l.keys("Y")
	if got := decodedClip(t, &out); got != "main.go:12: undefined: Foo" {
		t.Fatalf("Y copied the selected task's output as %q", got)
	}
	out.Reset()
	l.keys("esc", "Y")
	got := decodedClip(t, &out)
	if !strings.Contains(got, "kit:build │ ok example.com/kit") || !strings.Contains(got, "✖ 1/4 tasks succeeded") {
		t.Fatalf("with all output selected Y copied %q", got)
	}
}

func TestTheRunTabNamesHowToSelectTextAndSeeTheFullLogs(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	// m is how a plain drag selects text; the footer or the help lists it.
	l.keys("?")
	l.has("Copy the log shown", "Show the previous run", "Run the failed and skipped tasks again")
}

func TestTheExitSummaryNamesTheLastRunAndItsCommand(t *testing.T) {
	l := newLab(t, true, 140, 30)
	if got := l.sh().ExitSummary(); got != "" {
		t.Fatalf("nothing ran, summary %q", got)
	}
	l.keys("b")
	if got := l.sh().ExitSummary(); got != "✔ done in 0s · smith run build -p orders" {
		t.Fatalf("summary %q", got)
	}
	l.next = failingRun()
	l.keys("t")
	if got := l.sh().ExitSummary(); got != "✖ failed (code 1) in 0s · smith run test" {
		t.Fatalf("summary %q", got)
	}
}

func TestTheRunTabFillsTheTerminalWhileRunningAndAfterManyRuns(t *testing.T) {
	sizes := [][2]int{{140, 40}, {100, 30}, {99, 30}, {80, 24}, {72, 20}, {60, 15}, {40, 10}, {20, 5}, {10, 3}, {1, 1}}
	for _, sz := range sizes {
		l := newLab(t, true, sz[0], sz[1])
		l.next = failingRun()
		l.keys("b")
		l.next = &scripted{lines: []string{"kit:build ▶ x", "orders:build ▶ y"}, hold: true, release: make(chan struct{})}
		l.keys("t")
		check := func(when string) {
			lines := strings.Split(l.sh().View(), "\n")
			if len(lines) != sz[1] {
				t.Fatalf("%v %s: %d lines", sz, when, len(lines))
			}
			for i, line := range lines {
				if w := pane.Width(line); w != sz[0] {
					t.Fatalf("%v %s: line %d is %d cells wide: %q", sz, when, i, w, line)
				}
			}
		}
		check("while running")
		l.keys("ctrl+c")
		l.send(outputMsg{proc: l.sh().proc, exit: &runner.Exit{Code: 1, Cancelled: true}})
		check("after cancelling")
		l.keys(",")
		check("showing the earlier run")
		l.keys("?")
		check("with the help open")
	}
}

func TestTheTaskThatFailedOnItsOwnIsShownBeforeOneThatWasOnlyInterrupted(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = &scripted{lines: []string{
		"kit:build ▶ x", "orders:build ▶ y",
		"orders:build │ real reason", "orders:build ✖ exited with code 3",
	}, exit: runner.Exit{Code: 1}}
	l.keys("b")
	// kit:build was running, was seen first, and was interrupted by the end.
	l.has("Log: orders:build", "real reason")
}
