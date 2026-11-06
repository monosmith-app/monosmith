package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// splitKey handles the keys of the dividers: the backslash selects the next
// one and the bar the previous; Shift+arrows select one too (Shift+← and
// Shift+↑ the previous, Shift+→ and Shift+↓ the next, of the same direction);
// and, while one is selected, Tab goes on to the next one and Esc lets go. It reports whether the key was taken.
func (s *Shell) splitKey(name string) bool {
	if s.editing() || s.splitBlocked() {
		return false
	}
	spots := s.spots()
	in := splitter.IntentOf(name)
	switch in.Action {
	case splitter.CycleKey:
		next, ok := s.split.Cycle(spots, in.Dir)
		if !ok {
			return false
		}
		s.split = next
		s.say(words.Say("tui.split.selected"))
		return true
	case splitter.SelectKey:
		next, ok := s.split.Select(spots, in.Axis, in.Dir)
		if !ok {
			// No divider this way: the key keeps its other meaning (a list marks rows with it).
			return false
		}
		s.split = next
		s.say(words.Say("tui.split.selected"))
		return true
	case splitter.AdvanceKey:
		if s.split.SelectedIn(spots) {
			s.split, _ = s.split.Cycle(spots, in.Dir)
			return true
		}
	case splitter.ReleaseKey:
		if s.split.SelectedIn(spots) {
			s.split = s.split.Deselect()
			return true
		}
	}
	return false
}

// moveSplit moves the selected divider one step towards the right or the
// bottom (dir 1) or the left or the top (dir -1).
func (s *Shell) moveSplit(dir int) {
	next, p, ok := s.split.Move(s.spots(), dir, splitter.Step)
	if !ok {
		s.say(words.Say("tui.split.none"))
		return
	}
	s.split = next
	s.keepSplits()
	n, _ := s.split.Size(p.ID)
	s.say(words.Say("tui.split.moved", "size", n))
}

// splitMouse handles a press, a drag and a release on a divider. It reports
// whether the event was taken, so the click does not reach the panel beneath.
func (s *Shell) splitMouse(e tea.MouseEvent) bool {
	if s.splitBlocked() {
		return false
	}
	switch {
	case e.Action == tea.MouseActionPress && e.Button == tea.MouseButtonLeft:
		next, ok := s.split.Press(s.spots(), e.X, e.Y)
		s.split = next
		return ok
	case e.Action == tea.MouseActionMotion && s.split.Dragging != "":
		s.split = s.split.Drag(s.spots(), e.X, e.Y)
		s.applySplits()
		return true
	case e.Action == tea.MouseActionRelease && s.split.Dragging != "":
		s.split, _ = s.split.Release()
		s.keepSplits()
		return true
	}
	return false
}

// keepSplits gives the views the sizes and saves them, so the next start opens
// as the person left it.
func (s *Shell) keepSplits() {
	s.applySplits()
	if s.saveSplits != nil {
		s.saveSplits(s.split.Sizes)
	}
}
