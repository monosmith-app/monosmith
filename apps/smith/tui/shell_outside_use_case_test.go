package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// outsideLab is a shell with no workspace that can start commands in place,
// as chooseOnce gives it (sources.OutsideSources), with a fake process.
func outsideLab(t *testing.T, process *scripted) *lab {
	t.Helper()
	l := newLab(t, false, 120, 36)
	sh := l.sh()
	sh.sources = sources.Sources{
		Capture: func(_ context.Context, argv []string) sources.Captured {
			l.captured = append(l.captured, argv)
			return sources.Captured{OK: true, Text: "convert plan: 3 projects map, nothing needs you"}
		},
		Start: func(argv []string) (runner.Handle, error) {
			l.started = append(l.started, argv)
			return process, nil
		},
	}
	l.m = sh
	return l
}

func TestNewRunsInTheShellAndOffersToOpenTheWorkspace(t *testing.T) {
	l := outsideLab(t, &scripted{lines: []string{"Creating the workspace shop", "Wrote 14 file(s)"}})
	l.ask("new", "shop", "--ci", "github")
	l.has("smith new shop --ci github --yes")
	if len(l.started) != 0 {
		t.Fatalf("nothing runs before the answer: %v", l.started)
	}
	l.keys("enter")
	if got := l.lastStarted(); got != "new shop --ci github --yes" {
		t.Fatalf("started %q", got)
	}
	if l.sh().Argv != nil || l.quit {
		t.Fatalf("it runs in the shell, not in the terminal after it: %v", l.sh().Argv)
	}
	l.has("Done: ", "Wrote 14 file(s)", "Open the workspace here")
	l.keys("enter")
	if sh := l.sh(); !sh.reopen || sh.reopenDir != "shop" || !strings.Contains(sh.reopenNote, "first commit") {
		t.Fatalf("the shell opens again inside shop: reopen=%v dir=%q note=%q", sh.reopen, sh.reopenDir, sh.reopenNote)
	}
}

func TestAFailedNewShowsWhyAndBringsTheFormBackWithItsAnswers(t *testing.T) {
	l := outsideLab(t, &scripted{lines: []string{"smith: shop exists already"}, exit: runner.Exit{Code: 1}})
	l.m = func() Shell {
		sh := l.sh()
		sh.Form, sh.Answers = "Create a workspace", map[string][]string{"name": {"shop"}}
		return sh
	}()
	l.ask("new", "shop")
	l.keys("enter")
	l.has("It did not finish", "shop exists already")
	l.keys("enter")
	l.hasNot("It did not finish")
	if sh := l.sh(); sh.reopen {
		t.Fatal("a failure does not reopen the shell")
	}
}

func TestConvertShowsItsPlanFirstAndRunsInPlace(t *testing.T) {
	l := outsideLab(t, &scripted{lines: []string{"Converted 3 projects"}})
	l.ask("convert", "--commit")
	l.has("convert plan: 3 projects map")
	if len(l.started) != 0 {
		t.Fatalf("the plan comes first: %v", l.started)
	}
	l.keys("enter")
	if got := l.lastStarted(); got != "convert --commit --yes" {
		t.Fatalf("started %q", got)
	}
	l.has("Converted 3 projects")
	l.keys("enter")
	if sh := l.sh(); !sh.reopen || sh.reopenDir != "." {
		t.Fatalf("convert works in place: %q", sh.reopenDir)
	}
}

func TestAnythingElseOutsideAWorkspaceStillRunsInTheTerminal(t *testing.T) {
	l := outsideLab(t, &scripted{})
	l.ask("adopt", "--tags")
	l.keys("enter")
	if len(l.started) != 0 {
		t.Fatalf("adopt has its own wizard and the terminal: %v", l.started)
	}
}
