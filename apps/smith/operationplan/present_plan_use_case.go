package operationplan

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// ErrDeclined is returned when the person answers no to the plan.
var ErrDeclined = errors.New("cancelled: nothing was changed")

// Options say how a plan is presented.
type Options struct {
	// Yes is --yes: do not ask.
	Yes bool
	// DryRun is --dry-run: show the plan and change nothing.
	DryRun bool
	// Ask is false for a command that asks its own question.
	Ask bool
	// Env is the terminal state; DetectEnvironment gives the real one.
	Env Environment
}

// Present prints the plan and decides whether the command goes on. With
// DryRun it prints the plan and returns false. When ShouldAsk holds, it asks
// "Go ahead? [y/N]" and returns ErrDeclined for any answer but y. With --yes it
// returns true without waiting. Without a terminal (or in CI) and without
// --yes it REFUSES: nobody can answer, so smith does not decide for them. The
// error names --yes (apply) and --dry-run (see the plan).
//
// Parameters: out gets the plan and the question; in gives the answer; plan is
// what the command will do; o says how to present it.
// Returns: whether to write, and ErrDeclined when the person said no.
func Present(out io.Writer, in io.Reader, plan Plan, o Options) (bool, error) {
	Render(out, plan)
	if o.DryRun {
		fmt.Fprintln(out, "Dry run: nothing was changed.")
		return false, nil
	}
	if plan.Empty() || !o.Ask || o.Yes {
		return true, nil
	}
	if !ShouldAsk(o.Env, o.Yes) {
		return false, ErrNeedsAnswer()
	}
	fmt.Fprint(out, "Go ahead? [y/N] ")
	answer, _ := bufio.NewReader(in).ReadString('\n')
	if !strings.EqualFold(strings.TrimSpace(answer), "y") {
		return false, ErrDeclined
	}
	return true, nil
}

// ErrNeedsAnswer is the refusal for a run that has no terminal and no --yes.
// It states the problem, then the two ways out.
func ErrNeedsAnswer() error {
	return problem.New("nobody can confirm this change: there is no terminal and --yes was not given",
		"--yes applies the plan above", "--dry-run shows the plan and changes nothing")
}
