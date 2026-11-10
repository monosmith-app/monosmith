package completioncommand

import (
	"bufio"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The answers Ask returns.
const (
	AnswerYes   = "yes"
	AnswerNo    = "no"
	AnswerLater = "later"
	answerShow  = "show"
)

// maxScriptViews bounds how often "show me the script" is offered again.
const maxScriptViews = 3

// Terminal is the person at the keyboard: where the questions go and where
// the answers come from.
type Terminal struct {
	// In reads the answers.
	In *bufio.Reader
	// Out shows the questions.
	Out io.Writer
}

// Ask puts the plan in front of the person and asks whether to write it. The
// choice "Show me the script" prints the completion script and asks again.
//
// Parameters: t is the terminal; plan is what install would do; offer is true
// for the one-time offer, which adds "Ask me later" and "No, do not ask again"
// to the answers and takes "Ask me later" for Enter; false for the plain
// question of install, which takes "Yes" for Enter.
// Returns: AnswerYes, AnswerNo or AnswerLater.
func Ask(t Terminal, plan shellcompletion.Plan, offer bool) string {
	options := []choiceprompt.Option{
		{Label: words.Say("completion.ask.yes"), Value: AnswerYes, Hint: words.Say("completion.ask.yes.hint")},
		{Label: words.Say("completion.ask.no"), Value: AnswerNo, Hint: words.Say("completion.ask.no.hint")},
		{Label: words.Say("completion.ask.show"), Value: answerShow, Hint: words.Say("completion.ask.show.hint")},
	}
	prompt, def := words.Say("completion.ask.question"), AnswerYes
	if offer {
		options = []choiceprompt.Option{
			options[0],
			{Label: words.Say("completion.ask.never"), Value: AnswerNo, Hint: words.Say("completion.ask.never.hint")},
			{Label: words.Say("completion.ask.later"), Value: AnswerLater, Hint: words.Say("completion.ask.later.hint")},
			options[2],
		}
		prompt, def = words.Say("completion.offer.question", "shell", plan.Shell.Title()), AnswerLater
	}
	for range maxScriptViews {
		got := choiceprompt.Ask(prompt, options, def, t.In, t.Out)
		if got != answerShow {
			return got
		}
		for _, target := range plan.Targets {
			fmt.Fprintf(t.Out, "\n# %s\n%s\n", target.Path, strings.TrimSpace(target.Body))
		}
	}
	return def
}
