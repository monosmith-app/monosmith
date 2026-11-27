package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/userprefs"
)

// pinsSaver keeps the pinned projects of the workspace at root in the
// preferences file of the person (tui.json in the smith home), never in the
// workspace. Pins that cannot be written are lost at the next start.
func pinsSaver(root string) func(pins []string) {
	if root == "" {
		return nil
	}
	return func(pins []string) {
		home := pluginhost.Home()
		p, _ := userprefs.Load(home)
		if p.Pins == nil {
			p.Pins = map[string][]string{}
		}
		if len(pins) == 0 {
			delete(p.Pins, root)
		} else {
			p.Pins[root] = pins
		}
		_ = userprefs.Save(home, p)
	}
}
