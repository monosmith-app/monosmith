package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// unasked is a lab where the person has not said how smith treats undo.
func unasked(t *testing.T) *lab {
	l := newHomeLab(t, 120, 40)
	sh := l.sh()
	sh.commitChoice = ""
	l.m = sh
	return l
}

// change makes the shell run a command that changes the workspace.
func (l *lab) change(argv ...string) *lab {
	sh := l.sh()
	cmd := sh.execute(argv, false)
	l.m = sh
	l.drive(cmd)
	return l
}

func TestTheFirstChangeAsksHowSmithTreatsUndoAndRunsNothingYet(t *testing.T) {
	l := unasked(t).change("derive")
	l.has("Let smith commit each change so it can be undone?", "Yes, commit each change", "No, leave my files uncommitted", "Ask me each time")
	if len(l.started) != 0 || len(l.captured) != 0 {
		t.Fatalf("nothing may run before the answer: %v %v", l.started, l.captured)
	}
	l.keys("esc")
	l.hasNot("Let smith commit each change")
	if got := l.sh().commitChoice; got != "" {
		t.Fatalf("Esc saves nothing, got %q", got)
	}
}

func TestAskEachTimeIsSavedAndLeadsToTheUsualConfirmation(t *testing.T) {
	l := unasked(t)
	saved := ""
	sh := l.sh()
	sh.saveCommit = func(c string) { saved = c }
	l.m = sh
	l.change("derive").keys("down", "down", "enter")
	if saved != "ask" || l.sh().commitChoice != "ask" {
		t.Fatalf("saved %q, kept %q", saved, l.sh().commitChoice)
	}
	l.has("Commit this change so it can be undone")
	l.change("derive").hasNot("Let smith commit each change")
}

func TestYesWritesTheSettingAsACommitOfItsOwnThenGoesOn(t *testing.T) {
	l := unasked(t)
	l.answers(map[string]sources.Captured{"config set workspace.commit_operations true --commit": {OK: true}})
	l.change("derive").keys("enter")
	if len(l.captured) == 0 || len(l.captured[0]) < 5 || l.captured[0][4] != "--commit" {
		t.Fatalf("captured %v", l.captured)
	}
	if !l.sh().commitOps || !l.sh().commitSet {
		t.Fatal("undo is on after yes")
	}
	l.hasNot("Let smith commit each change")
}

func TestUndoThatIsOffExplainsAndOpensTheSameQuestion(t *testing.T) {
	l := newHomeLab(t, 120, 40)
	l.answers(map[string]sources.Captured{"undo --dry-run": {Text: "× smith has no operation to undo here"}})
	l.keys("ctrl+p").typeText("Undo the last").keys("enter")
	l.has("Undo is not on", "Choose how smith commits")
	l.keys("enter")
	l.has("Let smith commit each change so it can be undone?")
	if len(l.started) != 0 {
		t.Fatalf("nothing runs: %v", l.started)
	}
}
