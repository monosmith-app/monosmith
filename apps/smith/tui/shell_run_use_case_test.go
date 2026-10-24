package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

func failingRun() *scripted {
	return &scripted{lines: []string{
		"kit:build ▶ go build ./...",
		"orders:build ▶ go build ./...",
		"kit:build │ ok example.com/kit 0.3s",
		"kit:build ✔ 106ms",
		"orders:build │ main.go:12: undefined: Foo",
		"orders:build ✖ exited with code 2",
		"  ✖ orders:build exited with code 2",
		"  ↷ web:build because orders:build did not succeed",
		"sdk:test ✔ from cache",
		"✖ 1/4 tasks succeeded (0 from cache) in 1s",
	}, exit: runner.Exit{Code: 1}}
}

func rowOf(l *lab, text string) int {
	for i, line := range l.lines() {
		if strings.Contains(line, text) {
			return i
		}
	}
	return -1
}

func TestARunShowsEveryTaskItsStateAndAProgressBar(t *testing.T) {
	l := newLab(t, true, 140, 36)
	l.next = failingRun()
	l.keys("b")
	if l.lastStarted() != "run build -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l.has("$ smith run build -p orders", "failed", "4/4 100%", "orders:build", "exited with code 2", "kit:build", "passed", "106ms",
		"web:build", "skipped", "because orders:build did not succeed", "sdk:test")
	// Failures come first in the list.
	if rowOf(l, "orders:build  ") > rowOf(l, "kit:build  ") && rowOf(l, "kit:build  ") > 0 {
		t.Fatalf("a failure must be listed before a pass:\n%s", l.view())
	}
	// The failing task's log is shown without asking.
	l.has("Log: orders:build", "main.go:12: undefined: Foo")
	l.hasNot("ok example.com/kit")
	// The tab carries a badge, the status bar the result, the navigator a mark.
	l.has("Run ✖1")
}

func TestTheStatusBarShowsTheExactCommandWhileAndAfterItRuns(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.next = &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.keys("b")
	last := strings.Split(l.view(), "\n")
	status := last[len(last)-1]
	if !strings.Contains(status, "$ smith run build -p orders") || !strings.Contains(status, "running") {
		t.Fatalf("status bar: %q", status)
	}
}

func TestRunningTasksAreAnimatedByTicks(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = &scripted{lines: []string{"kit:build ▶ go build", "orders:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.keys("b")
	l.has("running", "0/4 0%", "Ctrl+C stop")
	l.hasNot("b build", "[■")
	first := l.sh().View()
	l.send(tickMsg{})
	second := l.sh().View()
	if first == second {
		t.Fatal("a tick must move the spinners")
	}
	frames := map[string]bool{}
	for i := 0; i < 10; i++ {
		l.send(tickMsg{})
		frames[l.sh().View()] = true
	}
	if len(frames) < 5 {
		t.Fatalf("only %d distinct frames in ten ticks", len(frames))
	}
	// The elapsed time counts up with the clock.
	l.advance(5 * 1e9)
	l.send(tickMsg{})
	l.has("5s")
}

func TestCtrlCStopsTheRunAndAsksTheProcessToStop(t *testing.T) {
	l := newLab(t, true, 140, 34)
	p := &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("b")
	l.keys("ctrl+c")
	if !p.stopped {
		t.Fatal("the process was not asked to stop")
	}
	l.has("Stopping")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("cancelled")
	l.hasNot("running")
	if l.quit {
		t.Fatal("one ctrl+c only stops the run")
	}
	// With nothing running, ctrl+c leaves.
	l.keys("ctrl+c")
	if !l.quit {
		t.Fatal("the next ctrl+c quits")
	}
}

func TestCtrlCTwiceWhileStoppingQuits(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.keys("b", "ctrl+c", "ctrl+c")
	if !l.quit {
		t.Fatal("a second ctrl+c quits")
	}
}

func TestACancelledRunLeavesNothingSpinning(t *testing.T) {
	l := newLab(t, true, 140, 34)
	p := &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("b", "ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	run := l.sh().run
	for _, task := range run.Tasks {
		if task.Status.String() == "running" || task.Status.String() == "queued" {
			t.Fatalf("%s is still %s", task.ID, task.Status)
		}
	}
}

func TestCachedTasksAreShownAsCached(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = &scripted{lines: []string{"kit:build ✔ from cache", "orders:build ✔ from cache", "web:build ✔ from cache", "sdk:test ✔ from cache", "✔ 4/4 tasks succeeded (4 from cache) in 10ms"}}
	l.keys("b")
	l.has("cached", "⚡", "4 cached", "passed")
	l.hasNot("failed")
}

func TestTheLogFollowsUntilYouScrollAndFTogglesIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	lines := []string{"kit:build ▶ x"}
	for i := 0; i < 120; i++ {
		lines = append(lines, fmt.Sprintf("kit:build │ line %d", i))
	}
	lines = append(lines, "kit:build ✔ 1s")
	l.next = &scripted{lines: lines}
	l.keys("b", "esc") // esc shows all output
	l.has("following", "line 119")
	l.keys("pgup")
	l.has("paused")
	l.hasNot("line 119")
	l.keys("end")
	l.has("following", "line 119")
	l.keys("f")
	l.has("paused")
	l.has("f follow")
	l.keys("f")
	l.has("following")
}

func TestRunAgainRepeatsTheLastCommand(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("t")
	l.keys("r")
	l.has("r again")
	if len(l.started) != 2 || l.started[0][0] != l.started[1][0] || l.lastStarted() != "run test -p orders" {
		t.Fatalf("started %v", l.started)
	}
}

func TestTheNavigatorMarksTheLastResultOfEachProject(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.next = failingRun()
	l.keys("b")
	_, y := l.find("orders go")
	if row := l.lines()[y]; !strings.Contains(row, "✖") {
		t.Fatalf("orders failed but its row says %q", row)
	}
	_, y = l.find("kit go")
	if row := l.lines()[y]; !strings.Contains(row, "✔") {
		t.Fatalf("kit passed but its row says %q", row)
	}
}

func TestEveryRunIsKeptInTheCommandLog(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("b")
	l.keys("t")
	l.keys("1", "2")
	if got := strings.Join(l.sh().history, " | "); got != "smith run build -p orders | smith run test -p orders" {
		t.Fatalf("history %q", got)
	}
}

func TestACommandThatCannotStartSaysSoInsteadOfHanging(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.sources.Start = func([]string) (runner.Handle, error) { return nil, fmt.Errorf("no such program") }
	l.m = sh
	l.keys("b")
	l.has("could not start: no such program")
	if l.sh().running() {
		t.Fatal("a failed start must not leave the run going")
	}
}

func TestProcessOutputAfterANewRunIsIgnored(t *testing.T) {
	l := newLab(t, true, 140, 34)
	old := &scripted{lines: []string{"kit:build ▶ x"}, hold: true, release: make(chan struct{})}
	l.next = old
	l.keys("b", "ctrl+c")
	l.send(outputMsg{proc: old, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.keys("t")
	l.send(outputMsg{proc: old, lines: []string{"kit:build ✖ stale"}})
	l.hasNot("stale")
}

func TestOutsideAWorkspaceACommandRunsInTheTerminalAfterTheShellQuits(t *testing.T) {
	sh := NewShell(Config{Version: "dev"})
	sh.execute([]string{"adopt"}, false)
	if !sh.quit || len(sh.Argv) == 0 || sh.Argv[0] != "adopt" {
		t.Fatalf("quit=%v argv=%v", sh.quit, sh.Argv)
	}
}
