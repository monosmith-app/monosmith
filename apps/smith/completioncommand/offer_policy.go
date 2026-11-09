package completioncommand

import (
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"time"
)

// askAgainAfter is how long "ask me later" waits.
const askAgainAfter = 14 * 24 * time.Hour

// ShouldAsk is true when the one-time offer may be shown now: the person did
// not say no, did not install it, and did not ask to wait past now.
//
// Parameters: a is the remembered answer; now is the time.
func ShouldAsk(a Answer, now time.Time) bool {
	switch a.Choice {
	case ChoiceNo, ChoiceInstalled:
		return false
	case ChoiceLater:
		return !now.Before(a.AskAfter)
	}
	return true
}

// HomeOfferWanted is true when the Home screen of the shell interface lists
// "Tab completion is not set up": nothing was installed, the person did not
// say no, the package of smith does not ship the file, and the shell is
// known. "Ask later" does not hide it: Home only waits.
//
// Parameters: env is the machine; a is the remembered answer.
func HomeOfferWanted(env shellenv.Environment, a Answer) bool {
	if a.Choice == ChoiceNo || a.Choice == ChoiceInstalled {
		return false
	}
	shell, ok := shellenv.DetectShell(env)
	if !ok || shellcompletion.PackagedPath(env, shell) != "" {
		return false
	}
	targets := shellcompletion.Targets(env, shell, "")
	if len(targets) == 0 {
		return false
	}
	for _, t := range targets {
		if shellcompletion.StateOf(env, t) != shellcompletion.Missing {
			return false
		}
	}
	return true
}
