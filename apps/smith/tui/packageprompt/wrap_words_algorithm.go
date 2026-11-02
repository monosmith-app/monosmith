package packageprompt

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// wrapWords breaks s into lines of at most w cells at spaces; a word longer
// than w is cut.
func wrapWords(s string, w int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		switch {
		case line == "":
			line = word
		case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	var out []string
	for _, l := range lines {
		out = append(out, wrap(l, w)...)
	}
	return out
}
