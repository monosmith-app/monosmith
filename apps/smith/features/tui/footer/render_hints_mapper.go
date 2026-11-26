package footer

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

// separator sits in the middle of the gap between two hints that follow each
// other, as in the short help of bubbles/help: "key desc • key desc".
const separator = "•"

// Render draws the placed hints in w cells; hover is the column of the pointer
// (-1 for none). Every hint is the key in the accent, in bold, and its
// description muted. Two hints one gap apart have a separator in the border
// colour between them. A hint a click presses is lit under the pointer.
//
// Parameters: th is the theme; placed are the hints from Place; hover is the
// pointer's column; w is the width of the row.
// Returns: the row, exactly w cells wide.
func Render(th theme.Theme, placed []Placed, hover, w int) string {
	var b strings.Builder
	x := 0
	for i, p := range placed {
		if i > 0 && p.X == x+gap {
			b.WriteString(" " + th.Border(separator, false) + " ")
		} else {
			b.WriteString(strings.Repeat(" ", max(p.X-x, 0)))
		}
		text := th.Accent(p.Hint.Glyph) + " " + th.Muted(p.Hint.Label)
		if p.Hint.Key != "" && hover >= p.X && hover < p.X+p.W {
			text = th.RowHover(text)
		}
		b.WriteString(text)
		x = p.X + p.W
	}
	return pane.Fit(b.String(), w)
}
