package pane

// MiddleCell is Cell for a name: one that is too long loses its middle, not
// its end, so the part that tells packages and projects apart (a scope at the
// start, a suffix at the end) stays readable.
func MiddleCell(s string, w int) string {
	return Cell(MiddleTruncate(s, w-1), w)
}
