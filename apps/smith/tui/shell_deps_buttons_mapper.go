package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// depsButtons is the line of buttons above the dependency list: the repair of
// the installed dependencies, which opens its own dialog. under is the id of
// the button under the pointer.
func (s Shell) depsButtons(w int, under string) []buttonRow {
	const id = "deps.repair"
	label := words.Say("tui.deps_repair.button")
	key := s.shortcutHere(id)
	state := s.buttonState(id, under)
	if c, ok := commands.ByID(id); ok && c.Unavailable(s.state()) != "" {
		state = button.Disabled
	}
	bw := button.Width(s.th, label, key)
	if bw > w {
		return nil
	}
	return []buttonRow{{line: button.Render(s.th, label, key, state), hit: hit{kind: "action", spans: []hitSpan{{from: 0, to: bw, id: id}}}}}
}
