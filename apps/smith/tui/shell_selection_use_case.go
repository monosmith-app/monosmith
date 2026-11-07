package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/selection"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
)

// selectionRegions are the panels a drag may select in: the inside of the main
// panel and of the detail beside it, cut apart along the dividers of the screen.
// The rule, the margins, the header, the tabs and the footer belong to none.
func (s Shell) selectionRegions() []layout.Rect {
	if s.overlaid() {
		return nil
	}
	areas := []layout.Rect{s.mainInner()}
	if s.lay().HasDetail() {
		areas = append(areas, s.detailInner())
	}
	var cuts []selection.Cut
	for _, p := range s.spots() {
		// A row division runs edge to edge, over the margins too: its cut is made
		// at a cell in the middle, which lies inside the panel it divides.
		cuts = append(cuts, selection.Cut{X: p.Hit.X + p.Hit.W/2, Y: p.Hit.Y, Vertical: p.Axis == splitter.Column})
	}
	return selection.Divide(areas, cuts)
}

// onButton is true when a press at (x, y) does something other than select: it
// runs an action or a step of the Run history.
func (s Shell) onButton(x, y int) bool {
	if in := s.mainInner(); in.Contains(x, y) && s.tab != commands.TabSettings {
		h, ok := s.mainContent(in.W, in.H).hits[y-in.Y]
		return ok && (h.kind == "action" || h.kind == "nav")
	}
	if in := s.detailInner(); s.lay().HasDetail() && in.Contains(x, y) {
		_, ok := s.detailContent(in.W, in.H).hits[y-in.Y]
		return ok
	}
	return false
}

// selectableAt is the region a drag that starts at (x, y) may use, and false when
// the press selects nothing: outside every panel, over a button, or under a dialog.
func (s Shell) selectableAt(x, y int) (layout.Rect, bool) {
	if s.onButton(x, y) {
		return layout.Rect{}, false
	}
	for _, r := range s.selectionRegions() {
		if r.Contains(x, y) {
			return r, true
		}
	}
	return layout.Rect{}, false
}

// pressAndSelect does what a press does, then lets a drag from it select text.
// A new press always drops the highlight of the one before.
func (s *Shell) pressAndSelect(x, y int) tea.Cmd {
	region, ok := s.selectableAt(x, y)
	s.selection = s.selection.Clear()
	cmd := s.click(x, y)
	if ok && !s.overlaid() {
		s.selection = s.selection.Press(region, x, y)
	}
	return cmd
}

// selectMouse follows a drag and ends it at the release. It reports whether the
// event was taken.
func (s *Shell) selectMouse(e tea.MouseEvent) bool {
	if !s.selection.Active {
		return false
	}
	switch {
	case e.Action == tea.MouseActionMotion && e.Button == tea.MouseButtonLeft:
		s.selection = s.selection.Drag(e.X, e.Y)
	case e.Action == tea.MouseActionMotion, e.Action == tea.MouseActionRelease:
		// A motion with no button held means the terminal lost the release.
		s.endSelection()
	default:
		return false
	}
	return true
}

// endSelection ends the drag: it copies the selected text and keeps the highlight
// until something else happens. A click that moved nothing selects nothing.
func (s *Shell) endSelection() {
	next, ok := s.selection.Release()
	s.selection = next
	if !ok {
		return
	}
	text := next.Text(s.baseRows())
	if text == "" {
		s.selection = s.selection.Clear()
		return
	}
	s.selected = text
	s.copyText(text)
}

// dropSelection handles a key while text is selected. Esc only drops the
// highlight and is taken; any other key drops it too and goes on. It reports
// whether the key was taken.
func (s *Shell) dropSelection(name string) bool {
	if !s.selection.Active && !s.selection.Held {
		return false
	}
	s.selection = s.selection.Clear()
	return name == "esc"
}

// keepSelection drops the highlight when the screen no longer shows the text it
// held at the release, or when a dialog covers it.
func (s *Shell) keepSelection() {
	if !s.selection.Held {
		return
	}
	if s.overlaid() || s.selection.Text(s.baseRows()) != s.selected {
		s.selection = s.selection.Clear()
	}
}

// paintSelection draws the highlight over the screen rows.
func (s Shell) paintSelection(rows []string) []string {
	if !s.selection.Selected() {
		return rows
	}
	return selection.Paint(s.th, rows, s.selection)
}
