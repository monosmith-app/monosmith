package rail

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// Render draws the rail without a frame: the title on the first row, a blank
// row, then each view on two rows, its key above its name, with a blank row
// between views. A shorter rail drops the blank rows, and a rail shorter
// still puts the key and the name on one row. A single rule in the Border
// colour runs down the last column. The open view has the mark "▶ " in the
// accent before its key, and its rows are on the Highlight background; the
// view under the pointer is on the hover background. The key row ends with
// the view's badge. Under the views, after a blank row, come the pinned
// projects under their muted bold heading: each one with its mark, its name
// and, when there is room, its kind, as many as the rail has rows for.
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
	inner, rows := p.Width-1, p.Height
	room := max(inner-2, 0)
	body := make([]string, rows)
	body[0] = pad(p.Theme.Title(pane.CutEllipsis(p.Title, max(room, 1))), room)
	step := 3
	switch {
	case rows >= 2+3*len(p.Views)-1:
	case rows >= 2+2*len(p.Views):
		step = 2
	default:
		step = 1
	}
	y := 2
	for _, v := range p.Views {
		if y >= rows {
			break
		}
		if step == 1 {
			body[y] = viewRow(p, v.ID, true, v.Key+" "+v.Label, v.Badge, room)
			out.Rows[y] = v.ID
			y++
			continue
		}
		body[y] = viewRow(p, v.ID, true, v.Key, v.Badge, room)
		out.Rows[y] = v.ID
		if y+1 < rows {
			body[y+1] = viewRow(p, v.ID, false, v.Label, "", room)
			out.Rows[y+1] = v.ID
		}
		y += step
	}
	if step == 3 {
		y--
	}
	pinned(p, body, out.Rows, y, step, room)
	rule := p.Theme.Border("│", false)
	for i, l := range body {
		if l == "" {
			l = strings.Repeat(" ", inner)
		}
		body[i] = l + rule
	}
	out.Lines = body
	return out
}

// pad is s fitted to room cells, with a blank cell on each side.
func pad(s string, room int) string { return " " + pane.Fit(s, room) + " " }

// viewRow is one row of the view id: text at the left, after the mark when
// marked, and badge at the right end, on the background of the view's state.
func viewRow(p Props, id string, marked bool, text, badge string, room int) string {
	th := p.Theme
	mark := ""
	if marked {
		mark = "  "
		if id == p.Active {
			mark = th.Accent("▶ ")
		}
	}
	right := 0
	if badge != "" {
		right = pane.Width(badge) + 1
	}
	left := mark + pane.CutEllipsis(text, max(room-pane.Width(mark)-right, 1))
	gap := max(room-pane.Width(left)-pane.Width(badge), 0)
	line := pane.Fit(left+strings.Repeat(" ", gap)+badge, room)
	switch id {
	case p.Active:
		line = th.Selected(line)
	case p.Hover:
		line = th.RowHover(line)
	}
	return " " + line + " "
}

// pinned draws the heading and the pinned projects into body from line from
// on: a blank line, the heading, a blank line when step leaves room for it,
// then one project per line while lines are left.
func pinned(p Props, body []string, rows map[int]string, from, step, room int) {
	if len(p.Pinned) == 0 {
		return
	}
	th := p.Theme
	y := from + 1
	if step == 1 {
		y = from
	}
	if y >= len(body)-1 {
		return
	}
	body[y] = pad(th.Muted(th.Bold(pane.CutEllipsis(p.PinnedTitle, max(room, 1)))), room)
	y++
	if step > 1 {
		y++
	}
	for _, pin := range p.Pinned {
		if y >= len(body) {
			return
		}
		left := pin.Mark + " " + pin.Label
		kind := ""
		if pane.Width(left)+1+pane.Width(pin.Kind) <= room {
			kind = th.Muted(pin.Kind)
		}
		if kind != "" {
			left = pane.CutEllipsis(left, max(room-pane.Width(kind)-1, 1))
		} else {
			left = pane.CutEllipsis(left, max(room, 1))
		}
		gap := max(room-pane.Width(left)-pane.Width(kind), 1)
		line := pane.Fit(left+strings.Repeat(" ", gap)+kind, room)
		if pin.ID == p.Hover {
			line = th.RowHover(line)
		}
		body[y] = " " + line + " "
		rows[y] = pin.ID
		y++
	}
}
