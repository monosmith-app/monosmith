package rail

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// Render draws the rail: a rounded frame with the title in its top border,
// and inside it one view per row with a blank row before each (none when the
// rail is too short for them). The open view has the mark "▶ " in the accent,
// in bold, and its row is on the Highlight background; the view under the
// pointer is on the hover background. Each row ends with the view's badge and
// its key.
//
// Parameters: p is what the rail is drawn from.
// Returns: the rail, exactly p.Height lines of p.Width cells, and the view of
// each line.
func Render(p Props) Rail {
	out := Rail{Rows: map[int]string{}}
	if p.Width < 4 || p.Height < 2 {
		out.Lines = pane.Block(nil, max(p.Width, 0), max(p.Height, 0))
		return out
	}
	inner, rows := p.Width-2, p.Height-2
	step := 2
	if rows < 2*len(p.Views)+1 {
		step = 1
	}
	body := make([]string, rows)
	for i, v := range p.Views {
		y := i*step + step - 1
		if y >= rows {
			break
		}
		body[y] = row(p, v, inner)
		out.Rows[y+1] = v.ID
	}
	out.Lines = pane.Frame(p.Theme, p.Title, body, p.Width, p.Height, false)
	return out
}

// row is one view in inner cells: a blank cell, the mark, the name, then the
// badge and the key at the right end, and a blank cell.
func row(p Props, v View, inner int) string {
	th := p.Theme
	mark := "  "
	if v.ID == p.Active {
		mark = th.Accent("▶ ")
	}
	right := v.Key
	if v.Badge != "" {
		right = v.Badge + "  " + v.Key
	}
	room := max(inner-2, 0)
	left := mark + pane.CutEllipsis(v.Label, max(room-pane.Width(right)-3, 1))
	gap := max(room-pane.Width(left)-pane.Width(right), 1)
	line := pane.Fit(left+strings.Repeat(" ", gap)+right, room)
	switch v.ID {
	case p.Active:
		line = th.Selected(line)
	case p.Hover:
		line = th.RowHover(line)
	}
	return " " + line + " "
}
