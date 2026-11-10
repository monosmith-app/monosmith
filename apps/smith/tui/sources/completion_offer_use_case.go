package sources

import (
	"github.com/monosmith-app/monosmith/apps/smith/completioncommand"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// CompletionOffer gives the shell the two things Home needs for the item "Tab
// completion is not set up": whether to list it, and how to remember "Not now".
// The answer lives in the preferences of the person (tui.json in the smith
// home), the same place the one-time offer of the command line writes to.
//
// Parameters: home is the smith home.
// Returns: wanted, which reads the files and the answer; decline, which keeps
// the answer "no" so that nothing offers completion again.
func CompletionOffer(home string) (wanted func() bool, decline func()) {
	memory := completioncommand.PreferencesMemory{Home: home}
	return func() bool {
			return completioncommand.HomeOfferWanted(shellenv.OSEnvironment{}, memory.Load())
		}, func() {
			_ = memory.Save(completioncommand.Answer{Choice: completioncommand.ChoiceNo})
		}
}
