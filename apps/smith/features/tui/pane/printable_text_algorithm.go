package pane

import "strings"

const (
	// tabWidth is how many cells a tab takes once Printable has expanded it.
	tabWidth = 4
	// lineFeed and escape are the control characters Printable keeps.
	lineFeed = 0x0a
	escape   = 0x1b
)

// Printable makes text safe to lay out by cell. A terminal moves the cursor on
// a carriage return and jumps to the next tab stop on a tab, while the width
// arithmetic counts both as nothing, so a line from a CRLF file or one indented
// with tabs is drawn over the wrong cells and breaks the box around it. Tabs
// become spaces and every other control character is dropped, except the
// escape character, which starts a colour code, and the line feed, which Wrap
// splits on.
func Printable(s string) string {
	if !strings.ContainsFunc(s, isUnsafe) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == 0x09:
			b.WriteString(strings.Repeat(" ", tabWidth))
		case isUnsafe(r):
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isUnsafe reports a control character that moves the cursor or draws nothing.
func isUnsafe(r rune) bool {
	return (r < 0x20 && r != escape && r != lineFeed) || r == 0x7f
}
