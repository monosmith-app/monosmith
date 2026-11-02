package pane

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Window draws the part of a list that fits in h lines: the lines from the
// scroll's top, with a scrollbar in the last column, and one line under them
// ("↑ 3 more · ↓ 12 more") that says how many rows are hidden on each side.
// It is the one rule for every list: the wheel scrolls (Scroll.By), a key that
// moves the selection sel keeps it in view, and the wheel never moves the
// selection.
//
// Parameters: rows are the lines of the whole list, already fitted to w-1
// cells (the bar takes the last one); sel is the line of the selection; sc
// holds the top between frames and may be nil.
// Returns: the lines to draw (a list that fits comes back as it is, with no
// bar and no indicator) and top, the index of the first list line shown.
func Window(th theme.Theme, rows []string, w, h, sel int, sc *Scroll) (out []string, top int) {
	if len(rows) <= h {
		return rows, 0
	}
	h = max(h-1, 1) // the indicator takes a line
	top = sc.Show(sel, h, len(rows))
	out = Scrollbar(th, rows[top:min(top+h, len(rows))], w, top, len(rows))
	above, below := top, len(rows)-top-h
	var note string
	switch {
	case above > 0 && below > 0:
		note = fmt.Sprintf("↑ %d more · ↓ %d more", above, below)
	case above > 0:
		note = fmt.Sprintf("↑ %d more", above)
	case below > 0:
		note = fmt.Sprintf("↓ %d more", below)
	}
	return append(out, "  "+th.Muted(note)), top
}
