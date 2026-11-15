package pane

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Cell fits s into a table cell of exactly w cells: text that is too long ends
// in an ellipsis instead of being cut mid-word, and the last cell is always
// left blank so the next column never runs into it.
func Cell(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if w == 1 {
		return " "
	}
	body := w - 1
	cut := s
	if ansi.StringWidth(s) > body {
		cut = ansi.Truncate(s, body, "…") + "\x1b[m"
	}
	if pad := w - ansi.StringWidth(cut); pad > 0 {
		cut += strings.Repeat(" ", pad)
	}
	return cut
}
