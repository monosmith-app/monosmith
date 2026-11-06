package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
)

// saveSplits keeps the sizes of the panels in the preferences file; a file that
// cannot be written loses only the sizes, so the error is dropped.
func saveSplits(sizes map[string]int) {
	home := pluginhost.Home()
	p, _ := userprefs.Load(home)
	p.Splits = sizes
	_ = userprefs.Save(home, p)
}
