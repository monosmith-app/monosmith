package splitter

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
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
			lines[y] = pane.Splice(lines[y], p.Hit.X, 1, cell)
		}
	case Row:
		if p.Hit.Y < len(lines) {
			lines[p.Hit.Y] = pane.Splice(lines[p.Hit.Y], p.Hit.X, p.Hit.W, style(strings.Repeat("━", p.Hit.W)))
		}
	}
	return lines
}

// PaintJunctions draws the junctions that touch the divider p in the colour of
// its look, so the divider and the tee that joins it to a rule read as one
// piece. The glyph stays the light one; only the colour changes. A Resting
// divider leaves the lines as they are. The lines are changed in place and
// returned.
func PaintJunctions(th theme.Theme, lines []string, js []Junction, p Spot, look Look) []string {
	if look == Resting {
		return lines
	}
	style := th.Accent
	if look == Dragging {
		style = th.ButtonActive
	}
	for _, j := range js {
		if j.Y < 0 || j.Y >= len(lines) || !p.touches(j) {
			continue
		}
		lines[j.Y] = pane.Splice(lines[j.Y], j.X, 1, style(j.Glyph))
	}
	return lines
}

// touches is true when the junction is on the divider or at either end of it.
func (p Spot) touches(j Junction) bool {
	if p.Axis == Column {
		return j.X == p.Hit.X && j.Y >= p.Hit.Y-1 && j.Y <= p.Hit.Y+p.Hit.H
	}
	return j.Y == p.Hit.Y && j.X >= p.Hit.X-1 && j.X <= p.Hit.X+p.Hit.W
}
