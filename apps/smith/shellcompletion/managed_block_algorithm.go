package shellcompletion

import "strings"

// The lines that mark what smith added to a file. A file that holds them was
// written by smith, and uninstall removes what lies between them.
const (
	beginMarker = "# >>> smith completion >>>"
	endMarker   = "# <<< smith completion <<<"
	// managedNote is the line inside the block that tells a reader where it
	// came from.
	managedNote = `# Managed by "smith completion install". Remove it with "smith completion uninstall".`
)

// splitLines splits text into lines without their line ends, and says whether
// the text used Windows line ends.
func splitLines(text string) (lines []string, crlf bool) {
	crlf = strings.Contains(text, "\r\n")
	lines = strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	return lines, crlf
}

// joinLines is the inverse of splitLines.
func joinLines(lines []string, crlf bool) string {
	if crlf {
		return strings.Join(lines, "\r\n")
	}
	return strings.Join(lines, "\n")
}

// blockSpan finds the marker lines in lines.
// Returns: the index of the begin line and of the end line, and false when
// the block is not there or is not closed.
func blockSpan(lines []string) (begin, end int, ok bool) {
	begin = -1
	for i, l := range lines {
		switch {
		case begin < 0 && strings.TrimSpace(l) == beginMarker:
			begin = i
		case begin >= 0 && strings.TrimSpace(l) == endMarker:
			return begin, i, true
		}
	}
	return 0, 0, false
}

// blockLines is the block as lines: the markers, the note and the body.
func blockLines(body string) []string {
	out := []string{beginMarker, managedNote}
	out = append(out, strings.Split(strings.TrimRight(strings.ReplaceAll(body, "\r\n", "\n"), "\n"), "\n")...)
	return append(out, endMarker)
}

// HasBlock is true when text holds the block smith writes.
func HasBlock(text string) bool {
	lines, _ := splitLines(text)
	_, _, ok := blockSpan(lines)
	return ok
}

// BlockBody reads the body between the markers of the block in text.
// Returns: the body without the note line, and false when there is no block.
func BlockBody(text string) (string, bool) {
	lines, _ := splitLines(text)
	b, e, ok := blockSpan(lines)
	if !ok {
		return "", false
	}
	inner := lines[b+1 : e]
	if len(inner) > 0 && inner[0] == managedNote {
		inner = inner[1:]
	}
	return strings.Join(inner, "\n"), true
}

// WithBlock puts the block in text. A block that is already there is replaced
// in place. Otherwise the block goes to the end after one blank line, and the
// rest of the text is not touched.
//
// Parameters: text is the file as it is (empty for a new file); body is what
// goes between the markers.
// Returns: the new text.
func WithBlock(text, body string) string {
	lines, crlf := splitLines(text)
	block := blockLines(body)
	if b, e, ok := blockSpan(lines); ok {
		out := append(append(append([]string{}, lines[:b]...), block...), lines[e+1:]...)
		return joinLines(out, crlf)
	}
	if strings.TrimSpace(text) == "" {
		return joinLines(append(block, ""), crlf)
	}
	if lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	return joinLines(append(lines, append(block, "")...), crlf)
}

// WithoutBlock takes the block out of text, with the one blank line that
// WithBlock put before it.
//
// Returns: the new text, and false when there was no block.
func WithoutBlock(text string) (string, bool) {
	lines, crlf := splitLines(text)
	b, e, ok := blockSpan(lines)
	if !ok {
		return text, false
	}
	if b > 0 && lines[b-1] == "" {
		b--
	}
	return joinLines(append(append([]string{}, lines[:b]...), lines[e+1:]...), crlf), true
}
