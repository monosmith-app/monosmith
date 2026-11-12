package button

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Cell is the text of a button before any style: " label (key) ", or " label "
// when key is "". Every clickable item in the TUI is drawn from it. The key is
// spelled the one way the footer and the help spell it ("enter" as "Enter",
// "ctrl+c" as "Ctrl+C"), whatever name the caller passes.
func Cell(label, key string) string {
	if key == "" {
		return " " + label + " "
	}
	return " " + label + " (" + commands.KeyLabel(key) + ") "
}

// Width is the cells a button takes. It is the same in every theme, so a click
// span counted without the theme still matches what is drawn; th is kept so a
// future theme can change that.
func Width(_ theme.Theme, label, key string) int { return pane.Width(Cell(label, key)) }

// Render draws a button in a state: a label with its key in parentheses on the
// background of the state. The key is the one that does the same as a click, so
// the screen shows every action and how to reach it. Without colour, brackets
// take the place of the two padding cells and reverse video marks the other
// states, so the width never changes with the state or the theme. A Disabled
// button is muted text.
func Render(th theme.Theme, label, key string, state State) string {
	cell := Cell(label, key)
	if !th.Tokens.Colour {
		cell = "[" + cell[1:len(cell)-1] + "]"
	}
	switch state {
	case Hover:
		return th.ButtonHover(cell)
	case Active:
		return th.ButtonActive(cell)
	case Disabled:
		return th.Muted(cell)
	}
	return th.Button(cell)
}
