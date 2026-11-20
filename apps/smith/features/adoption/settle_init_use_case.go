package adoption

import (
	"bufio"
	"errors"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/projecttooling"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
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
	// Tooling are the --tooling answers (with --level full, smith's for every
	// project that has no answer of its own).
	Tooling projecttooling.Answers
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
	// Tooling is the answer of each Node project, by package name: its own
	// build and lint tools, or smith's.
	Tooling map[string]projecttooling.Answer
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
	interactive := s.Env.Terminal && !s.Env.CI
	tooling, toolingErr := DecideTooling(r, s.Tooling, interactive, s.Yes, false, s.In, s.Out)
	// smith's tooling replaces the project's own ESLint setup, so the move
	// puts that setup aside rather than asking about it.
	if s.Move.EslintSetup == "" && tooling[rootProjectName(r)] == projecttooling.Smith {
		s.Move.EslintSetup = "shared"
	}
	s.Move.In, s.Move.Out, s.Move.Yes = s.In, s.Out, s.Yes
	c, plan, moveErr := ChooseMove(r, s.Move)
	switch {
	case toolingErr != nil && moveErr == nil:
		moveErr = toolingErr
	case toolingErr != nil:
		moveErr = mergeProblems(moveErr, toolingErr)
	}
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
	return SettledInit{Candidate: c, Move: plan, Choices: choices, Tooling: tooling}, nil
}

// DecideTooling assesses the Node projects of the report (the project at the
// root too, before it moves) and settles, for each, whether it keeps its own
// build and lint tools or uses smith's (projecttooling.Decide).
//
// Parameters: r is the report; given the --tooling answers; interactive,
// yes and preview say who answers; in and out are the terminal.
// Returns: the answers by package name; an error when nobody can answer.
func DecideTooling(r Report, given projecttooling.Answers, interactive, yes, preview bool, in *bufio.Reader, out io.Writer) (map[string]projecttooling.Answer, error) {
	var dirs []string
	if r.RootProject != nil && r.RootProject.Ecosystem == adoptmove.EcosystemNode {
		dirs = append(dirs, ".")
	}
	for _, p := range r.Projects {
		if p.Language == "node" && p.Root != "." {
			dirs = append(dirs, p.Root)
		}
	}
	var projects []projecttooling.Assessment
	for _, d := range dirs {
		a, err := projecttooling.Assess(r.Root, d)
		if err != nil {
			return nil, err
		}
		projects = append(projects, a)
	}
	return projecttooling.Decide(projecttooling.DecideRequest{
		Projects: projects, Given: given, Interactive: interactive, Yes: yes, Preview: preview, In: in, Out: out,
	})
}

// rootProjectName is the package name of the Node project at the repository
// root, "" when there is none.
func rootProjectName(r Report) string {
	if r.RootProject == nil {
		return ""
	}
	a, err := projecttooling.Assess(r.Root, ".")
	if err != nil || !a.Node {
		return ""
	}
	return a.Name
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
