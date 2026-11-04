package pane

import (
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Blank empties the text of rows from..to-1 and keeps the frame, so a modal
// drawn over them has no half words showing at its edges. The first and last
// row of the range are frame edges and keep every box-drawing cell. In the rows
// between, only the cells in the columns keep are kept (the left and right
// border and the divider between panes); a box-drawing glyph the content drew
// elsewhere (a log rule, a scrollbar thumb) is blanked, or it would show as a
// stray fragment around the modal. Rows outside the range are left as they are.
//
// Parameters: rows are the screen lines; from and to bound the rows (to is
// exclusive); keep are the columns of the vertical frame lines.
// Returns: a copy of rows with the range emptied.
func Blank(rows []string, from, to int, keep ...int) []string {
	out := append([]string(nil), rows...)
	from = max(from, 0)
	to = min(to, len(out))
	for y := from; y < to; y++ {
		edge := y == from || y == to-1
		var b strings.Builder
		col := 0
		for _, r := range ansi.Strip(out[y]) {
			// A wide cell takes two columns, so it becomes two spaces.
			w := max(ansi.StringWidth(string(r)), 0)
			if r >= 0x2500 && r <= 0x257f && (edge || slices.Contains(keep, col)) {
				b.WriteRune(r)
			} else {
				b.WriteString(strings.Repeat(" ", w))
			}
			col += w
		}
		out[y] = b.String()
	}
	return out
}
