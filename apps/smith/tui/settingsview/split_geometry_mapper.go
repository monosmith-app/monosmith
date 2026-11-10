package settingsview

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// WithSideDetail says whether the shell shows the detail of the selected row in
// a panel beside the tab. The tab then draws no detail of its own and gives its
// rows the room.
//
// Parameters: beside is true when that panel is on screen.
// Returns: the tab with the choice.
func (m Model) WithSideDetail(beside bool) Model {
	m.sideDetail = beside
	return m
}

// DetailLines tell the selected row for the panel beside the tab: what it is,
// what it does and which key changes it, in lines of at most w cells. They are
// empty while there is no row.
//
// Parameters: th styles the lines; w is the width of the panel.
// Returns: the lines.
func (m Model) DetailLines(th theme.Theme, w int) []string {
	if !m.loaded && m.section.readsWorkspace() {
		return nil
	}
	list := m.items(m.section)
	cursor := m.cursor[m.section]
	if cursor >= len(list) {
		return nil
	}
	return m.detail(th, list[cursor], w)
}

// Identity changes when the selected row changes, so the panel beside the tab
// starts at its top for each row.
func (m Model) Identity() string {
	return fmt.Sprint(int(m.section), "|", m.cursor[m.section])
}

// Geometry is where the tab's two dividers are, in the cells of the tab.
type Geometry struct {
	// ListW is the width of the list of sections, between ListMin and ListMax.
	// The divider is the column just after it.
	ListW, ListMin, ListMax int
	// SepRow is the row of the divider between the list and the detail of the
	// right pane, or -1 when the pane has no detail now. DetailRows is the rows
	// of the detail, between DetailMin and DetailMax.
	SepRow, DetailRows, DetailMin, DetailMax int
}

// WithSizes gives the list of sections a width and the detail a number of
// rows; 0 for either is automatic.
func (m Model) WithSizes(listW, detailRows int) Model {
	m.listW, m.detailRows = listW, detailRows
	return m
}

// Geometry is where the dividers are for a tab of w by h cells.
func (m Model) Geometry(w, h int) Geometry {
	g := Geometry{ListW: m.sectionsWidth(w), ListMin: minSections, ListMax: max(w-minRight, minSections), SepRow: -1}
	if w < 30 || m.ed != nil || !m.loaded {
		g.ListMax = g.ListW
		return g
	}
	listH, detailH, _ := m.rowSplit(h - 1)
	if detailH > 0 {
		g.SepRow, g.DetailRows, g.DetailMin = 1+listH, detailH, minDetailRows
		g.DetailMax = max(h-1-1-5, minDetailRows)
	}
	return g
}
