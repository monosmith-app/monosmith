package pane

import "github.com/monosmith-app/monosmith/apps/smith/tui/theme"

// Scrollbar draws a scrollbar in the last column of rows, the rows of a list
// that shows total rows from row top on: a faint line with a brighter thumb
// whose length is the part of the list in view and whose place is where that
// part is. A list that fits has none, and its rows come back as they are, so
// the bar is there exactly when something is hidden.
//
// Each row keeps its own width w; its last cell is replaced by the bar's.
func Scrollbar(th theme.Theme, rows []string, w, top, total int) []string {
	h := len(rows)
	if h == 0 || total <= h || w < 2 {
		return rows
	}
	size := max(h*h/total, 1)
	span := total - h
	at := 0
	if span > 0 {
		at = (min(max(top, 0), span)*(h-size) + span/2) / span
	}
	out := make([]string, h)
	for i, row := range rows {
		glyph := th.Faint(th.Icons.Track)
		if i >= at && i < at+size {
			glyph = th.Accent(th.Icons.Thumb)
		}
		out[i] = Fit(row, w-1) + glyph
	}
	return out
}
