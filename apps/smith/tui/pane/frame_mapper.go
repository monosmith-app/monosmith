package pane

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Frame draws body inside a rounded border w cells wide and h rows high, with
// title set into the top edge. A focused pane has an accent border; the weight is the
// same for every pane and dialog, only the colour changes. The result has exactly h lines of w cells.
func Frame(th theme.Theme, title string, body []string, w, h int, focused bool) []string {
	if w < 2 || h < 2 {
		return Block(nil, w, h)
	}
	inner := w - 2
	label := ""
	if title != "" {
		label = " " + title + " "
		if focused {
			label = th.Accent(label)
		} else {
			label = th.Muted(label)
		}
	}
	dashes := max(inner-Width(label)-1, 0)
	top := th.Border("╭─", focused) + label + th.Border(strings.Repeat("─", dashes)+"╮", focused)
	if Width(label)+1 > inner {
		top = th.Border("╭"+strings.Repeat("─", inner)+"╮", focused)
	}
	rows := make([]string, 0, h)
	rows = append(rows, Fit(top, w))
	side := th.Border("│", focused)
	content := Block(body, inner, h-2)
	for _, line := range content {
		rows = append(rows, side+line+side)
	}
	rows = append(rows, th.Border("╰"+strings.Repeat("─", inner)+"╯", focused))
	return rows
}
