package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
)

// minBodyRows is the least rows the scrolling part of a split detail keeps,
// whatever the fixed part wants.
const minBodyRows = 4

// splitsDetail is true while the detail of a project has two parts: the facts
// of the project, which never scroll, above the actions, which do. A narrow
// screen opens the detail over the list in one scrolling piece, as before.
func (s Shell) splitsDetail() bool {
	if !s.inWorkspace || s.tab != commands.TabProjects || s.lay().Mode != layout.Wide || s.runShown() {
		return false
	}
	_, ok := s.list.Selected()
	return ok
}

// splitPanel cuts a panel in the fixed part (its sections) and the part that
// scrolls (its actions and the sections under them).
func splitPanel(p detail.Panel) (top, body detail.Panel) {
	top = detail.Panel{Sections: p.Sections, ListVisible: p.ListVisible}
	p.Sections = nil
	return top, p
}

// detailSplit draws the fixed part of the detail in a w by h area, with a thin
// rule under it, and says how many rows are left for the part that scrolls.
// Without a split there is no fixed part and the body has all h rows.
func (s Shell) detailSplit(w, h int) (top []string, bodyH int) {
	if !s.splitsDetail() {
		return nil, h
	}
	panel, _ := splitPanel(s.selectionPanel())
	panel.ListVisible = !s.detailFull() && s.lay().HasDetail()
	lines, _ := panel.View(s.th, w, "")
	if room := max(h-minBodyRows-1, 0); len(lines) > room {
		lines = lines[:room]
	}
	// The same thin rule as the one over the run part of a split panel.
	top = append(lines, " "+s.th.Faint(strings.Repeat("─", max(w-2, 1))))
	return top, max(h-len(top), 1)
}
