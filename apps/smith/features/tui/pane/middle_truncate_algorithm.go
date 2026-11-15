package pane

import "github.com/charmbracelet/x/ansi"

// MiddleTruncate shortens s to at most w cells by replacing its middle with an
// ellipsis, so a long path keeps both its start and its file name. A string
// that fits is returned unchanged.
func MiddleTruncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	total := ansi.StringWidth(s)
	if total <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	tail := (w - 1) / 2
	head := w - 1 - tail
	return ansi.Truncate(s, head, "") + "…" + ansi.TruncateLeft(s, total-tail, "")
}
