package pane

import "github.com/charmbracelet/x/ansi"

// Splice replaces the w cells of line from column x with s and keeps the
// styles of the rest. A column past the end of the line leaves it as it is.
//
// Parameters: line is one screen row; x and w are the first column and the
// number of cells to replace; s is the text put there.
// Returns: the new row.
func Splice(line string, x, w int, s string) string {
	total := ansi.StringWidth(line)
	if x >= total || x < 0 {
		return line
	}
	return ansi.Cut(line, 0, x) + s + ansi.Cut(line, min(x+w, total), total)
}
