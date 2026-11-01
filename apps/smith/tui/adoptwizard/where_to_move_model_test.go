package adoptwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
)

func rootProjectFlow(app bool) *fakeFlow {
	return &fakeFlow{found: adoptflow.Found{
		Kind: adoptflow.KindPlain, Source: "repository", Name: "solo", NoFull: adoptflow.NoFullForPlain,
		RootProject: &adoptmove.Candidate{Ecosystem: adoptmove.EcosystemNode, Manifest: "package.json", Name: "web", App: app},
	}}
}

func TestAProjectAtTheRootIsAskedWhereItGoesAndNothingIsAssumed(t *testing.T) {
	w := start(t, rootProjectFlow(true))
	w.want("is at the repository root", "the next screen asks where to move it")
	w.press("enter")
	if w.m.Stage() != StageMove {
		t.Fatalf("stage %v, want the question", w.m.Stage())
	}
	w.want("Move it into a folder of its own so smith can manage it?",
		"1) apps/web  (suggested)", "2) packages/web", "3) libs/web", "4) another folder", "5) no, leave it", "Name: web", "e edits the name")
	w.press("enter")
	if w.m.Stage() != StageReview {
		t.Fatalf("stage %v, want the review", w.m.Stage())
	}
	w.want("$ smith adopt --init --move apps/web", "[ ] Commit this change")
	w.press("enter")
	if w.m.Stage() != StageDone {
		t.Fatalf("stage %v", w.m.Stage())
	}
}

func TestALibraryAtTheRootSuggestsAPackageFolder(t *testing.T) {
	w := start(t, rootProjectFlow(false)).press("enter")
	w.want("1) packages/web  (suggested)", "2) apps/web")
	w.press("enter")
	w.want("$ smith adopt --init --move packages/web")
}

func TestTheNameIsEditedWithE(t *testing.T) {
	w := start(t, rootProjectFlow(true)).press("enter")
	w.press("e", "backspace", "backspace", "backspace", "u", "i")
	w.want("Type the name")
	// The name was "web": three Backspaces empty it, then ui is typed.
	if got := w.screen(); !strings.Contains(got, "Name: ui") || !strings.Contains(got, "1) apps/ui") {
		t.Fatalf("the edit is not shown:\n%s", got)
	}
	// While editing, a digit is text and Esc stops the editing, not the dialog.
	w.press("2", "esc")
	if w.m.Stage() != StageMove || !strings.Contains(w.screen(), "Name: ui2") {
		t.Fatalf("stage %v:\n%s", w.m.Stage(), w.screen())
	}
	w.press("e", "backspace", "enter")
	w.press("2")
	w.want("$ smith adopt --init --move packages/ui")
}

func TestAnUnusableNameAsksAgainWithTheReason(t *testing.T) {
	w := start(t, rootProjectFlow(true)).press("enter")
	w.press("e", "X", "enter", "1")
	if w.m.Stage() != StageMove {
		t.Fatalf("stage %v, want the question again", w.m.Stage())
	}
	w.want("! project name", "must be lowercase")
}

func TestAnotherFolderTakesTheWholeText(t *testing.T) {
	w := start(t, rootProjectFlow(true)).press("enter")
	w.press("e", "backspace", "backspace", "backspace", "t", "o", "o", "l", "s", "/", "w", "enter", "4")
	w.want("$ smith adopt --init --move tools/w")
}

func TestLeavingTheProjectPassesNone(t *testing.T) {
	w := start(t, rootProjectFlow(true)).press("enter", "5")
	w.want("$ smith adopt --init --move none")
}

func TestBackFromTheReviewReturnsToTheQuestionAndFromTheQuestionToWhatWasFound(t *testing.T) {
	w := start(t, rootProjectFlow(true)).press("enter", "enter")
	w.press("b")
	if w.m.Stage() != StageMove {
		t.Fatalf("stage %v, want the question", w.m.Stage())
	}
	w.press("b")
	if w.m.Stage() != StageFound {
		t.Fatalf("stage %v, want what was found", w.m.Stage())
	}
}
