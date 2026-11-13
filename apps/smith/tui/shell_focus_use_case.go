package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// beside is true while the open view shows its detail (or the run) in a panel
// beside the list, so the keyboard has two panels to choose from.
func (s Shell) beside() bool {
	return s.lay().HasDetail() && !s.detailFull() && (s.tab == commands.TabProjects || s.tab == commands.TabDoctor)
}

// sideKeys is true while the keys go to the right panel: the detail or the run
// that fills the pane, or the one beside the list after Tab or Right moved the
// keyboard there.
func (s Shell) sideKeys() bool { return s.detailFull() || (s.sideFocus && s.beside()) }

// sideNext is true where Tab moves the keyboard to the right panel next: from
// the list when no run is kept, and from the recent runs.
func (s Shell) sideNext() bool {
	if !s.beside() || s.sideFocus {
		return false
	}
	return s.tab == commands.TabDoctor || s.recentFocus || len(s.runs) == 0
}

// focusSide moves the keyboard to the right panel when it is beside the list.
func (s *Shell) focusSide() {
	if s.beside() {
		s.sideFocus, s.recentFocus = true, false
	}
}

// focusList gives the keyboard back to the list of projects (or of checks).
func (s *Shell) focusList() { s.sideFocus, s.recentFocus = false, false }

// keepDetailCursorVisible scrolls the detail just enough to show the action the
// keyboard cursor is on, its button and the lines that say what it does, under
// no "more" marker. When both do not fit, the button wins.
func (s *Shell) keepDetailCursorVisible() {
	actions := s.detailActions()
	if s.detailCursor < 0 || s.detailCursor >= len(actions) {
		return
	}
	r := s.detailRect()
	_, bodyH := s.detailSplit(r.W, r.H)
	// Drawn with the cursor mark, as the panel is: the mark narrows the notes.
	panel := s.detailPanel()
	panel.Cursor = s.detailCursorID(panel)
	// A run part pinned under the body is always in view: only the body scrolls.
	panel, bodyH, _ = s.pinRun(panel, r.W, bodyH)
	lines, rows := panel.View(s.th, r.W, "")
	if s.detailCursor == 0 && !s.detailOnRun {
		// The first action: show the panel from its top, with the headings above it.
		s.detailTop = 0
		return
	}
	// In a split detail the run button is shown with the action it runs, so the
	// description between them shows too.
	want := map[string]bool{panel.Cursor: true}
	if panel.Splits() {
		want[detail.RunID] = true
		if s.detailOnRun {
			want = map[string]bool{detail.RunID: true}
		}
	}
	first, last := -1, -1
	for row, id := range rows {
		if want[id] {
			if first < 0 || row < first {
				first = row
			}
			last = max(last, row)
		}
	}
	if first < 0 {
		return
	}
	top := pane.KeepVisible(s.detailTop, last, len(lines), bodyH)
	s.detailTop = pane.KeepVisible(top, first, len(lines), bodyH)
}
