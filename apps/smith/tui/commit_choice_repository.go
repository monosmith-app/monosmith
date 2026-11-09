package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
)

// commitChoiceSaver keeps the answer "no" or "ask" to the commit question for
// the workspace at root, in the preferences file of the person. An answer that
// cannot be written is lost, and smith asks again at the next start.
func commitChoiceSaver(root string) func(choice string) {
	if root == "" {
		return nil
	}
	return func(choice string) {
		home := pluginhost.Home()
		p, _ := userprefs.Load(home)
		if p.CommitChoices == nil {
			p.CommitChoices = map[string]string{}
		}
		p.CommitChoices[root] = choice
		_ = userprefs.Save(home, p)
	}
}
