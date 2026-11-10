package footer

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Render draws the placed hints in w cells; hover is the column of the pointer
// (-1 for none). A hint a click presses is a button, lit under the pointer; the
// others are text.
func Render(th theme.Theme, placed []Placed, hover, w int) string {
	var b strings.Builder
	x := 0
	for _, p := range placed {
		b.WriteString(strings.Repeat(" ", max(p.X-x, 0)))
		text := th.Bold(p.Hint.Glyph) + " " + th.Muted(p.Hint.Label)
		if p.Hint.Key != "" {
			state := button.Rest
			if hover >= p.X && hover < p.X+p.W {
				state = button.Hover
			}
			text = button.Render(th, p.Hint.Label, p.Hint.Glyph, state)
		}
		b.WriteString(text)
		x = p.X + p.W
	}
	return pane.Fit(b.String(), w)
}
