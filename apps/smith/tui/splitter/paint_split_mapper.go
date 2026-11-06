package splitter

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Paint draws a divider in its look over the screen lines (one string per row,
// in screen coordinates): a heavy rule in the accent while selected, and in the
// active button colour while held. A Resting divider leaves the lines as they are.
// The lines are changed in place and returned.
func Paint(th theme.Theme, lines []string, p Spot, look Look) []string {
	if look == Resting {
		return lines
	}
	style := th.Accent
	if look == Dragging {
		style = th.ButtonActive
	}
	switch p.Axis {
	case Column:
		cell := style("┃")
		for y := p.Hit.Y; y < p.Hit.Y+p.Hit.H && y < len(lines); y++ {
			lines[y] = splice(lines[y], p.Hit.X, 1, cell)
		}
	case Row:
		if p.Hit.Y < len(lines) {
			lines[p.Hit.Y] = splice(lines[p.Hit.Y], p.Hit.X, p.Hit.W, style(strings.Repeat("━", p.Hit.W)))
		}
	}
	return lines
}

// splice replaces the w cells of a line from column x with s.
func splice(line string, x, w int, s string) string {
	total := ansi.StringWidth(line)
	if x >= total {
		return line
	}
	return ansi.Cut(line, 0, x) + s + ansi.Cut(line, min(x+w, total), total)
}
