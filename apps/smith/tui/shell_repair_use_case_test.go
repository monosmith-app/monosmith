package tui

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// repairLab is a shell over a smith.toml that uses a renamed key. applied
// counts the writes the shell asked for.
func repairLab(t *testing.T) (*lab, *int) {
	t.Helper()
	applied := 0
	r := &sources.ConfigRepair{
		Fix: doctorfix.Planned{Fix: protocol.Fix{Description: "rename the keys smith.toml uses to their current names", Safe: true,
			Files: []protocol.FileEdit{{Path: "smith.toml", Content: "[[groups]]\n"}}}},
		Show: func(doctorfix.Planned) (string, error) {
			return "-[[task_groups]]\n+[[groups]]\n", nil
		},
		Apply: func(context.Context, protocol.Fix, io.Writer) error { applied++; return nil },
	}
	sh := NewShell(Config{
		Theme: theme.Default(), Problem: "smith.toml uses [[task_groups]], which is now [[groups]]",
		Repair: r, Actions: guide.FixConfigFirst(nil),
	})
	l := &lab{t: t, clock: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), m: sh}
	l.send(tea.WindowSizeMsg{Width: 120, Height: 30})
	return l, &applied
}

func TestFixSmithTomlIsTheFirstEntryAndWritesNothingBeforeTheUserConfirms(t *testing.T) {
	l, applied := repairLab(t)
	l.has("Fix smith.toml", "Enter on Fix smith.toml shows the change")
	if strings.Contains(l.view(), "run smith again") {
		t.Error("the dead-end message stays when the fix is offered")
	}
	l.keys("enter") // opens the review
	l.has("[[task_groups]]", "[[groups]]")
	if *applied != 0 {
		t.Fatal("the file was written before the user confirmed")
	}
	l.keys("esc")
	if *applied != 0 || l.sh().fix != nil {
		t.Fatal("Esc cancels the review and writes nothing")
	}
}

func TestFixSmithTomlAppliesOnConfirmAndReloadsTheWorkspace(t *testing.T) {
	l, applied := repairLab(t)
	l.keys("enter", "enter") // review, then confirm
	if *applied != 1 {
		t.Fatalf("the fix was applied %d times, want 1", *applied)
	}
	if sh := l.sh(); !sh.reopen || !sh.quit || !strings.Contains(sh.reopenNote, "Fixed smith.toml") {
		t.Fatalf("the shell does not start again over the fixed file: reopen=%v quit=%v note=%q", sh.reopen, sh.quit, sh.reopenNote)
	}
}

func TestAnUnfixableSmithTomlKeepsTheOldMessage(t *testing.T) {
	sh := NewShell(Config{Theme: theme.Default(), Problem: "smith.toml:3:9: expected value"})
	if got := strings.Join(sh.brokenConfigBanner(120), " "); !strings.Contains(got, "Fix that line, then run smith again") {
		t.Errorf("the message changed: %q", got)
	}
}
