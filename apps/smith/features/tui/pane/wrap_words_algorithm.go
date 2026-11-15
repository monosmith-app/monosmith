package pane

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// WrapWords breaks s into lines of at most w cells at spaces. Each paragraph
// (text between newlines) wraps on its own, and a word longer than w is cut
// hard. A width under 1 returns s as one line.
func WrapWords(s string, w int) []string {
	if w < 1 {
		return []string{s}
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			for ansi.StringWidth(word) > w {
				if line != "" {
					out, line = append(out, line), ""
				}
				out = append(out, ansi.Truncate(word, w, ""))
				word = ansi.TruncateLeft(word, w, "")
			}
			switch {
			case line == "":
				line = word
			case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
				line += " " + word
			default:
				out, line = append(out, line), word
			}
		}
		out = append(out, line)
	}
	return out
}

// WrapHard cuts s into lines of exactly w cells (the last may be shorter),
// whatever the content: spaces count as any other cell. A width under 1, or a
// string that fits, returns s as one line.
func WrapHard(s string, w int) []string {
	if w < 1 || ansi.StringWidth(s) <= w {
		return []string{s}
	}
	var out []string
	for ansi.StringWidth(s) > w {
		out = append(out, ansi.Truncate(s, w, ""))
		s = ansi.TruncateLeft(s, w, "")
	}
	return append(out, s)
}
