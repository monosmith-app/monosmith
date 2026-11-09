package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// actionAt is the action whose button is drawn at the screen cell (x, y), in the
// main pane or in the detail beside it, and "" when the cell is on none. Home has
// its own buttons, Settings has none, and a dialog hides every button below it.
// This one lookup serves the hover look, the pressed look and the click, so a
// button never shows one state and acts on another.
func (s Shell) actionAt(x, y int) string {
	if s.probing || s.overlaid() || s.tab == commands.TabHome {
		return ""
	}
	s.probing = true
	if in := s.mainInner(); in.Contains(x, y) {
		if s.tab == commands.TabSettings {
			return ""
		}
		if h, ok := s.mainContent(in.W, in.H-s.toastRows(in.H)).hits[y-in.Y]; ok && h.kind == "action" {
			if len(h.spans) > 0 {
				// A line of several buttons: the one under the cell.
				sp, _ := h.at(x - in.X)
				return sp.id
			}
			return h.id
		}
		return ""
	}
	if in := s.detailInner(); s.lay().HasDetail() && in.Contains(x, y) {
		if h, ok := s.sideContent(in.W, in.H-s.toastRows(in.H)).hits[y-in.Y]; ok && h.kind == "action" {
			if len(h.spans) > 0 {
				// A line of several buttons: the one under the cell.
				sp, _ := h.at(x - in.X)
				return sp.id
			}
			return h.id
		}
	}
	return ""
}

// buttonState is how the button of action id looks, given under, the action the
// pointer is over: held down while the mouse button is pressed on it, hovered
// otherwise, at rest when the pointer is elsewhere.
func (s Shell) buttonState(id, under string) button.State {
	switch {
	case id != under:
		return button.Rest
	case s.pressed == id:
		return button.Active
	}
	return button.Hover
}

// pressButton starts a press on a button: the button shows its active look and
// nothing runs until the mouse button is let go over the same one. It reports
// whether the press was on a button, so a press elsewhere goes on as before.
func (s *Shell) pressButton(x, y int) bool {
	id := s.actionAt(x, y)
	if id == "" {
		return false
	}
	s.selection = s.selection.Clear()
	s.pressed = id
	return true
}

// releaseButton ends a press on a button. The action runs when the pointer is
// still over the button that was pressed; letting go anywhere else cancels it.
func (s *Shell) releaseButton(x, y int) tea.Cmd {
	id := s.pressed
	s.pressed = ""
	if id == "" || s.actionAt(x, y) != id {
		return nil
	}
	return s.click(x, y)
}
