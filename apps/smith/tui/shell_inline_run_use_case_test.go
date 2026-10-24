package tui

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

// inlineLab swaps the terminal-owning process for a script.
func inlineLab(t *testing.T, code int) (*lab, *[]*runner.Inline) {
	t.Helper()
	l := newLab(t, true, 140, 30)
	var made []*runner.Inline
	sh := l.sh()
	sh.sources.Inline = func(argv []string) *runner.Inline {
		c := runner.NewInlineWith(argv, func(_ io.Reader, o, _ io.Writer) (int, error) {
			_, _ = o.Write([]byte("building\r\n"))
			return code, nil
		}, nil).WithClock(func() time.Time { return l.clock })
		made = append(made, c)
		return c
	}
	l.m = sh
	return l, &made
}

func TestATaskRunUsesTheTerminalInsteadOfTheRunTabLog(t *testing.T) {
	l, made := inlineLab(t, 0)
	l.keys("b")
	if len(*made) != 1 || strings.Join((*made)[0].Argv, " ") != "run build -p orders" {
		t.Fatalf("inline runs: %d", len(*made))
	}
	if len(l.started) != 0 {
		t.Fatalf("the captured runner must not start: %v", l.started)
	}
	// Bubble Tea would hand over the terminal here; the script stands in for it.
	var out bytes.Buffer
	c := (*made)[0]
	c.SetStdout(&out)
	_ = c.Run()
	l.send(inlineDoneMsg{cmd: c})
	l.has("Runs", "run build -p orders", "passed")
	if got := l.sh().ExitSummary(); got != "✔ done in 0s · smith run build -p orders" {
		t.Fatalf("summary %q", got)
	}
}

func TestAFailedInlineRunIsListedAndRepeatable(t *testing.T) {
	l, made := inlineLab(t, 2)
	l.keys("b")
	c := (*made)[0]
	_ = c.Run()
	l.send(inlineDoneMsg{cmd: c})
	l.has("failed (exit 2)")
	l.keys("r")
	if len(*made) != 2 || strings.Join((*made)[1].Argv, " ") != "run build -p orders" {
		t.Fatalf("r must run the last command again, got %d runs", len(*made))
	}
}

func TestWithoutAnInlineSourceTheRunStaysCaptured(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("b")
	if l.lastStarted() != "run build -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
}
