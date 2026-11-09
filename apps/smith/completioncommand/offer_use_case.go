package completioncommand

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"time"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// OfferOptions is what Offer needs.
type OfferOptions struct {
	// Env is the machine.
	Env shellenv.Environment
	// Memory keeps the answer.
	Memory Memory
	// Terminal is where the question is asked.
	Terminal Terminal
	// Script makes the completion script of a shell.
	Script func(shellenv.Shell) (string, error)
	// Now is the time; the zero value means time.Now.
	Now time.Time
}

// Offer asks once, after the first interactive run, whether to set up tab
// completion. The caller has checked that a person is at the terminal and that
// this is not CI. Offer shows the exact files, then asks: Yes writes them, No
// is remembered and never asked again, Ask me later waits two weeks.
//
// It does nothing when the person answered before, when the completion is
// already installed or shipped by the package, or when the shell is not known.
// A failure to write is reported as an error and is not remembered, so the
// offer comes back.
//
// Parameters: o is the machine, the memory and the terminal.
// Returns: an error when the script or a file cannot be written.
func Offer(o OfferOptions) error {
	now := o.Now
	if now.IsZero() {
		now = time.Now()
	}
	answer := o.Memory.Load()
	if !ShouldAsk(answer, now) || !HomeOfferWanted(o.Env, Answer{}) {
		return nil
	}
	shell, _ := shellenv.DetectShell(o.Env)
	script, err := o.Script(shell)
	if err != nil {
		return err
	}
	plan := shellcompletion.PlanFor(o.Env, shell, script)
	fmt.Fprintf(o.Terminal.Out, "\n%s\n", shellcompletion.RenderPlan(plan))
	switch Ask(o.Terminal, plan, true) {
	case AnswerYes:
		changes, err := shellcompletion.Install(o.Env, plan)
		if err != nil {
			return err
		}
		shellcompletion.PrintChanges(o.Terminal.Out, plan.Shell, changes)
		return o.Memory.Save(Answer{Choice: ChoiceInstalled})
	case AnswerNo:
		fmt.Fprintln(o.Terminal.Out, words.Say("completion.offer.declined"))
		return o.Memory.Save(Answer{Choice: ChoiceNo})
	}
	fmt.Fprintln(o.Terminal.Out, words.Say("completion.offer.later"))
	return o.Memory.Save(Answer{Choice: ChoiceLater, AskAfter: now.Add(askAgainAfter)})
}
