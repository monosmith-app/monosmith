package button

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// RowSpans places a row of buttons side by side, one cell apart, centred in
// the inner width of a dialog.
//
// Parameters: labels and keys are what each button shows (see Render); inner
// is the width inside the dialog's border and padding.
// Returns: for each button, the columns [from, to) it covers, counted from the
// dialog's left edge. The row starts two cells in, for the border and the
// padding, and is never placed left of that.
func RowSpans(th theme.Theme, labels, keys []string, inner int) [][2]int {
	widths := make([]int, len(labels))
	total := max(len(labels)-1, 0)
	for i, label := range labels {
		widths[i] = Width(th, label, keys[i])
		total += widths[i]
	}
	x := 2 + max((inner-total)/2, 0)
	spans := make([][2]int, len(labels))
	for i, w := range widths {
		spans[i] = [2]int{x, x + w}
		x += w + 1
	}
	return spans
}

// At is the button under the cell (x, y) of a dialog placed at rect, whose
// buttons sit on the second row from the bottom (the row above the border).
// It returns the index into spans, or -1 when the cell is on no button.
func At(spans [][2]int, rect layout.Rect, x, y int) int {
	if y != rect.Y+rect.H-2 {
		return -1
	}
	for i, s := range spans {
		if x-rect.X >= s[0] && x-rect.X < s[1] {
			return i
		}
	}
	return -1
}
