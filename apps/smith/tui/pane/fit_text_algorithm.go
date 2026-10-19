package pane

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Fit pads or cuts a single line to exactly w cells. It never wraps.
func Fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	cut := ansi.Truncate(s, w, "")
	if cut != s {
		cut += "\x1b[m" // a style that was cut short must not run on into the padding
	}
	if pad := w - ansi.StringWidth(cut); pad > 0 {
		cut += strings.Repeat(" ", pad)
	}
	return cut
}

// Block returns exactly h lines of exactly w cells, padding with blank lines.
func Block(lines []string, w, h int) []string {
	out := make([]string, max(h, 0))
	for i := range out {
		if i < len(lines) {
			out[i] = Fit(lines[i], w)
		} else {
			out[i] = Fit("", w)
		}
	}
	return out
}

// Width is the number of cells a string takes.
func Width(s string) int { return ansi.StringWidth(s) }

// Overlay draws over on top of base starting at cell (x, y); both are blocks
// of equal-width lines. Cells under over are replaced.
func Overlay(base []string, over []string, x, y int) []string {
	out := append([]string(nil), base...)
	for i, line := range over {
		row := y + i
		if row < 0 || row >= len(out) {
			continue
		}
		w := ansi.StringWidth(line)
		left := ansi.Truncate(out[row], x, "")
		left += strings.Repeat(" ", max(x-ansi.StringWidth(left), 0))
		right := ansi.TruncateLeft(out[row], x+w, "")
		out[row] = left + "\x1b[m" + line + "\x1b[m" + right
	}
	return out
}

// Cut shortens s to at most w cells without padding it.
func Cut(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "")
}

// Strip removes the colour codes from s.
func Strip(s string) string { return ansi.Strip(s) }
