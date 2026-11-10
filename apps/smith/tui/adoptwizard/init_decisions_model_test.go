package adoptwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
)

func plainFlow(facts adoptdecision.Facts) *fakeFlow {
	return &fakeFlow{facts: facts, found: adoptflow.Found{
		Kind: adoptflow.KindPlain, Source: "repository", Name: "solo", NoFull: adoptflow.NoFullForPlain,
		Projects: []adoptflow.Project{{Name: "solo", Root: ".", Language: "go", Kind: "app"}},
	}}
}

// The wizard asks the decisions adopt --init makes, says what each answer does,
// and puts the safe answer first. Nothing is written before the review.
func TestTheWizardAsksTheDecisionsOfAdoptInitBeforeTheReview(t *testing.T) {
	f := plainFlow(adoptdecision.Facts{Detected: []string{"gitlab"}, OwnPipelines: map[string]string{"gitlab": ".gitlab-ci.yml"}})
	w := start(t, f).press("enter")
	if w.m.Stage() != StageInit {
		t.Fatalf("stage %v", w.m.Stage())
	}
	w.want("Line endings", "No, keep my line endings", "nothing is renormalised")
	if strings.Contains(strings.Join(f.calls, "|"), "apply") {
		t.Fatalf("nothing is written while the decisions are asked: %v", f.calls)
	}
	w.press("enter") // the safe answer
	w.want("Which CI should smith write", "gitlab", "the repository already uses it")
	w.press("enter")
	w.want("You already have a pipeline: .gitlab-ci.yml", "yours moves to")
	w.press("enter")
	if w.m.Stage() != StageReview {
		t.Fatalf("stage %v, want the review", w.m.Stage())
	}
	// The review shows the plan of adopt --init, with its diffs.
	w.want("Plan: adopt smith", "create smith.toml")
}

// The answers become flags of the step, so --yes takes them and no default of its own.
func TestTheAnswersReachTheAdoptionAsFlags(t *testing.T) {
	f := plainFlow(adoptdecision.Facts{})
	w := start(t, f).press("enter")
	w.press("2")     // yes, write .gitattributes
	w.press("enter") // the CI the repository uses (github)
	if w.m.Stage() != StageReview {
		t.Fatalf("stage %v", w.m.Stage())
	}
	if want, calls := "adopt --gitattributes add --agent skip --ci github --ci-existing replace", strings.Join(f.calls, "|"); !strings.Contains(calls, want) {
		t.Fatalf("the adoption step lacks %q: %v", want, f.calls)
	}
}

// Back from the review returns to the decisions, and from them to what was found.
func TestBackFromTheReviewReturnsToTheDecisions(t *testing.T) {
	w := start(t, plainFlow(adoptdecision.Facts{})).press("enter").init()
	w.press("b")
	if w.m.Stage() != StageInit {
		t.Fatalf("stage %v, want the decisions", w.m.Stage())
	}
	w.press("b")
	if w.m.Stage() != StageFound {
		t.Fatalf("stage %v, want what was found", w.m.Stage())
	}
}
