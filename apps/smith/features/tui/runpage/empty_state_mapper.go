package runpage

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Keys are the keys the empty state names: build, test, the palette and the
// Projects view.
type Keys struct {
	Build, Test, Palette, Projects string
}

// Empty draws the Runs view before anything ran in this session: a title, then
// the ways to start a run, each with its key.
//
// Parameters: th is the theme; k are the keys; w is the width.
// Returns: the lines, each at most w cells.
func Empty(th theme.Theme, k Keys, w int) []string {
	lines := []string{
		th.Title(words.Say("runpage.empty.title")),
		"",
		th.Muted(words.Say("runpage.empty.intro")),
		"",
		words.Say("runpage.empty.build", "key", th.Accent(k.Build)),
		words.Say("runpage.empty.test", "key", th.Accent(k.Test)),
		words.Say("runpage.empty.palette", "key", th.Accent(k.Palette)),
		words.Say("runpage.empty.projects", "key", th.Accent(k.Projects)),
		"",
		th.Faint(words.Say("runpage.empty.session")),
	}
	var out []string
	for _, l := range lines {
		for _, part := range pane.Wrap(l, max(w-1, 4), 0) {
			out = append(out, pane.Fit(" "+part, w))
		}
	}
	return out
}
