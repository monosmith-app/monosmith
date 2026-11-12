package dashboard

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// maxRecent is how many runs the line names.
const maxRecent = 3

// runGap is the cells between two runs of the line.
const runGap = 3

// maxRunCommand is the cells a command may take before it ends in an ellipsis.
const maxRunCommand = 28

// recent draws the last runs on one line: a status mark, the command and how
// long it took, each a target that opens the run. A run that does not fit in
// the width is left out, and the first one is cut short rather than left out.
func (c *canvas) recent(runs []Run) {
	th := c.th
	lead := " " + th.Muted(words.Say("dashboard.recent.label")) + strings.Repeat(" ", runGap)
	line, x, row := lead, pane.Width(lead), len(c.lines)
	for i, r := range runs[:min(len(runs), maxRecent)] {
		took := ""
		if r.Took != "" {
			took = " " + r.Took
		}
		command := shortCommand(r.Command)
		room := c.w - 1 - x - pane.Width(took) - 2
		if i > 0 && pane.Width(command) > room {
			break
		}
		command = pane.CutEllipsis(command, min(maxRunCommand, max(room, 1)))
		cell := c.runIcon(r) + " " + command + th.Muted(took)
		cw := pane.Width(c.runIcon(r)) + 1 + pane.Width(command) + pane.Width(took)
		if i > 0 {
			line += strings.Repeat(" ", runGap)
			x += runGap
		}
		if len(c.targets) == c.sel {
			// The selected look of every pick list: the cursor mark, in the gap
			// before the run, and an accent label.
			mark := choicerow.Cursor(th, true)
			line = ansi.Truncate(line, x-pane.Width(mark), "") + th.Accent(mark)
			cell = th.Accent(pane.Strip(cell))
		}
		c.targets = append(c.targets, Target{Row: row, From: x, To: x + cw, Run: r.Index})
		line += cell
		x += cw
	}
	c.add(line)
}

// runIcon is the status mark of a run that ended.
func (c *canvas) runIcon(r Run) string {
	switch {
	case r.Cancelled:
		return c.th.Muted(c.th.Icons.Cancelled)
	case r.OK:
		return c.th.Success(c.th.Icons.Passed)
	}
	return c.th.Failure(c.th.Icons.Failed)
}
