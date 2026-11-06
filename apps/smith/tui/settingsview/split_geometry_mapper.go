package settingsview

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
