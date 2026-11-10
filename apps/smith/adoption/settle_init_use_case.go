package adoption

import (
	"bufio"
	"errors"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// InitSettings are the answers and the terminal state that smith adopt --init
// must settle before it changes anything.
type InitSettings struct {
	// Move is what --move said.
	Move MoveOptions
	// Flags are the --gitattributes, --agent, --ci and --ci-existing answers.
	Flags adoptdecision.Flags
	// Yes is --yes.
	Yes bool
	// Install is --install, for the plan.
	Install bool
	// Env is the terminal state.
	Env operationplan.Environment
	// In is the one reader of every question; Out gets the questions.
	In  *bufio.Reader
	Out io.Writer
}

// SettledInit is everything --init decided before the first change.
type SettledInit struct {
	// Candidate and Move are nil when the root project stays where it is.
	Candidate *adoptmove.Candidate
	Move      *adoptmove.Plan
	// Choices are the settled answers of the other questions.
	Choices adoptdecision.Choices
}

// SettleInit answers every open question of adopt --init and checks that
// someone can confirm, all before the first file changes. A refusal (no
// terminal, a missing flag, or a missing --yes) names every missing flag in one
// message, and nothing was moved or written when it happens.
//
// Parameters: r is the inspection report; s the answers and the terminal.
// Returns: the move and the choices, or an error that names the flags to give.
func SettleInit(r Report, s InitSettings) (SettledInit, error) {
	var out SettledInit
	s.Move.In, s.Move.Out, s.Move.Yes = s.In, s.Out, s.Yes
	c, plan, moveErr := ChooseMove(r, s.Move)
	interactive := s.Env.Terminal && !s.Env.CI
	if moveErr != nil && interactive && !s.Yes {
		return out, moveErr
	}
	w := s.Out
	if moveErr != nil {
		w = io.Discard
	}
	choices, err := adoptdecision.Resolve(adoptdecision.Request{
		Flags: s.Flags, Facts: InitFacts(r), Yes: s.Yes,
		Interactive: interactive, In: s.In, Out: w,
	})
	switch {
	case moveErr != nil && err != nil:
		return out, mergeProblems(moveErr, err)
	case moveErr != nil:
		return out, moveErr
	case err != nil:
		return out, err
	}
	if err := operationplan.NeedsAnswerFor(PlanInit(r, choices, s.Install), operationplan.Options{Yes: s.Yes, Ask: true, Env: s.Env}); err != nil {
		return out, err
	}
	return SettledInit{Candidate: c, Move: plan, Choices: choices}, nil
}

// rollbackTarget is the git revision that holds the tree as it was before the move.
func rollbackTarget(moveCommitted bool) string {
	if moveCommitted {
		return "HEAD~1"
	}
	return "HEAD"
}

// mergeProblems joins two refusals into one message with every remedy.
func mergeProblems(a, b error) error {
	var pa, pb *problem.Error
	if !errors.As(a, &pa) || !errors.As(b, &pb) {
		return a
	}
	return problem.New("nobody can answer the questions of smith adopt --init: there is no terminal and --yes was not given",
		append(append([]string{pa.Problem}, pa.Remedy...), pb.Remedy...)...)
}
