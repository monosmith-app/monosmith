package settingsview

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

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

// WithSizes gives the list of sections a width and the detail a number of
// rows; 0 for either is automatic.
func (m Model) WithSizes(listW, detailRows int) Model {
	m.listW, m.detailRows = listW, detailRows
	return m
}
