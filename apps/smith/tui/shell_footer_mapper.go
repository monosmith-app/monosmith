package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/footer"
)

// footerHints are the key hints of the open view, placed in the screen's width.
func (s Shell) footerHints() []footer.Placed {
	view, tail := commands.Hints(s.state())
	return footer.Place(view, tail, s.w)
}

// footerRow draws the hints; the one under the pointer is highlighted.
func (s Shell) footerRow() string {
	hover := -1
	if s.mouseY == s.lay.Footer.Y {
		hover = s.mouseX
	}
	return footer.Render(s.th, s.footerHints(), hover, s.w)
}

// namedKeys are the keys of the footer that are not one character.
var namedKeys = map[string]tea.KeyType{
	"enter": tea.KeyEnter, "esc": tea.KeyEsc, "tab": tea.KeyTab, "shift+tab": tea.KeyShiftTab,
	"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
	"pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown, "home": tea.KeyHome, "end": tea.KeyEnd,
	"ctrl+c": tea.KeyCtrlC, "ctrl+u": tea.KeyCtrlU, "ctrl+p": tea.KeyCtrlP, "f1": tea.KeyF1, "delete": tea.KeyDelete,
}

// pressedKey is the key message a click on a footer hint stands for.
func pressedKey(name string) tea.KeyMsg {
	if t, ok := namedKeys[name]; ok {
		return tea.KeyMsg{Type: t}
	}
	if name == " " {
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}
