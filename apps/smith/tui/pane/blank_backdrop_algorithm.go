package pane

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Blank empties the text of rows from..to-1 and keeps their frame lines
// (the box-drawing cells), so a modal drawn over them has no half words
// showing at its edges. Rows outside the range are left as they are.
func Blank(rows []string, from, to int) []string {
	out := append([]string(nil), rows...)
	for y := max(from, 0); y < min(to, len(out)); y++ {
		var b strings.Builder
		for _, r := range ansi.Strip(out[y]) {
			if r >= 0x2500 && r <= 0x257f {
				b.WriteRune(r)
				continue
			}
			// A wide cell takes two columns, so it becomes two spaces.
			b.WriteString(strings.Repeat(" ", max(ansi.StringWidth(string(r)), 0)))
		}
		out[y] = b.String()
	}
	return out
}
