package choicerow

// Span is the lines [From, To) one row takes on screen, counted from the first
// line of the body it is drawn in. A row that wraps takes several lines.
type Span struct{ From, To int }

// At is the index of the row whose span holds line, or -1 when line is on no
// row. Every pick list maps a click and a hover to a row with it, so a press
// anywhere on a wrapped row counts.
//
// Parameters: spans are the rows in order; line is a body line.
// Returns: the row index, or -1.
func At(spans []Span, line int) int {
	for i, sp := range spans {
		if line >= sp.From && line < sp.To {
			return i
		}
	}
	return -1
}
