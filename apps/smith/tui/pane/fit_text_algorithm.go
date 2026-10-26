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
	s = Printable(s)
	cut := ansi.Truncate(s, w, "")
	if cut != s {
		cut += "\x1b[m" // a style that was cut short must not run on into the padding
	}
	if pad := w - ansi.StringWidth(cut); pad > 0 {
		cut += strings.Repeat(" ", pad)
	}
	return cut
}

// Wrap breaks s into lines of at most w cells at word boundaries (a word
// longer than w is split), keeping colours; every line after the first is
// indented by indent cells.
func Wrap(s string, w, indent int) []string {
	if w <= indent+1 {
		return []string{Cut(s, w)}
	}
	parts := strings.Split(ansi.Wrap(Printable(s), w-indent, "/\\"), "\n")
	pad := strings.Repeat(" ", indent)
	for i := 1; i < len(parts); i++ {
		parts[i] = pad + strings.TrimLeft(parts[i], " ")
	}
	return parts
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
func Width(s string) int { return ansi.StringWidth(Printable(s)) }

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
	return ansi.Truncate(Printable(s), w, "")
}

// Strip removes the colour codes from s.
func Strip(s string) string { return ansi.Strip(s) }
