package adoptwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
)

func splitFlow(plan adoptmove.Plan) *fakeFlow {
	f := rootProjectFlow(true)
	f.plan = plan
	return f
}

// The wizard asks about each development dependency smith cannot place and
// about the ESLint setup, with the answer that changes the least first, and
// passes the answers to the step as flags.
func TestTheWizardAsksWhatSmithCannotPlaceAndPassesTheAnswers(t *testing.T) {
	w := start(t, splitFlow(adoptmove.Plan{Unsure: []string{"supertest", "nock"}, Lint: true})).press("enter", "enter")
	if w.m.Stage() != StageSplit {
		t.Fatalf("stage %v, want the split questions", w.m.Stage())
	}
	w.want("1 of 3", "supertest is a development dependency, and smith does not know it as a tool",
		"The project: it stays in the project's own manifest", "The workspace: it moves to the workspace's manifest")
	// supertest goes to the workspace, nock stays, ESLint is replaced by the shared config.
	w.press("2")
	w.want("nock is a development dependency")
	w.press("1")
	w.want("The project has its own ESLint setup", "Keep my own ESLint setup in the project", "Use the shared config")
	w.press("2").init()
	w.want("--hoist supertest", "--eslint-setup shared")
	if strings.Contains(w.screen(), "nock") && strings.Contains(w.screen(), "--hoist supertest,nock") {
		t.Fatalf("nock stays in the project:\n%s", w.screen())
	}
}

func TestTheFirstAnswerKeepsEverythingInTheProject(t *testing.T) {
	w := start(t, splitFlow(adoptmove.Plan{Unsure: []string{"supertest"}, Lint: true})).press("enter", "enter")
	w.press("1", "1").init()
	w.want("--eslint-setup keep")
	if strings.Contains(w.screen(), "--hoist") {
		t.Fatalf("no dependency was sent to the workspace:\n%s", w.screen())
	}
}

func TestNothingToAskSkipsTheSplitAndLeavingTheProjectNeverAsks(t *testing.T) {
	w := start(t, splitFlow(adoptmove.Plan{})).press("enter", "enter")
	if w.m.Stage() != StageInit {
		t.Fatalf("stage %v, want the decisions of adopt --init", w.m.Stage())
	}
	w = start(t, splitFlow(adoptmove.Plan{Unsure: []string{"supertest"}, Lint: true})).press("enter", "5")
	if w.m.Stage() != StageInit {
		t.Fatalf("stage %v: a project that stays at the root is not split", w.m.Stage())
	}
}

func TestBackFromTheDecisionsReturnsToTheSplitAndFromTheSplitToTheMove(t *testing.T) {
	w := start(t, splitFlow(adoptmove.Plan{Unsure: []string{"supertest"}})).press("enter", "enter", "1")
	if w.m.Stage() != StageInit {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.press("b")
	if w.m.Stage() != StageSplit {
		t.Fatalf("stage %v, want the split questions", w.m.Stage())
	}
	w.press("b")
	if w.m.Stage() != StageMove {
		t.Fatalf("stage %v, want where it goes", w.m.Stage())
	}
}
