package runhistory

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Window is the rows of the list of recent runs that fit in rows lines, the
// chosen one (at, from 0) in view, out of total runs: first is the position of
// the first row and n how many rows there are.
func Window(at, total, rows int) (first, n int) {
	n = min(rows, total)
	return max(at-n+1, 0), n
}

// Row is one run of the list of recent runs, as plain text.
type Row struct {
	Pointer, Command, Result, Took, Ago string
	Selected                            bool
}

// Line draws a row in w cells: the command, how it ended, how long it took and
// how long ago. muted styles the age and selected the whole chosen row; both
// come from the theme.
func Line(r Row, w int, muted, selected func(string) string) string {
	cmdW := max(w-3-1-10-1-8-1-10, 12)
	// A command too long for its column ends with an ellipsis, as every cut line does.
	line := " " + r.Pointer + " " + pane.Fit(pane.CutEllipsis(r.Command, cmdW), cmdW) + " " + pane.Fit(r.Result, 10) + " " +
		fmt.Sprintf("%8s", r.Took) + " " + muted(fmt.Sprintf("%10s", r.Ago))
	if r.Selected {
		return selected(pane.Fit(pane.Strip(line), w))
	}
	return pane.Fit(line, w)
}
