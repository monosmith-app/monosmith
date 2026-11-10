package selection

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Paint draws the selected cells over the screen lines (one string per row, in
// screen coordinates) in the theme's pointer-over colour: a bright background in
// a colour theme, reverse video in the plain one, so the highlight is seen with
// no colour at all. It changes no cell width. The lines are changed in place
// and returned.
func Paint(th theme.Theme, lines []string, s State) []string {
	for y := range lines {
		from, to, ok := s.cols(y)
		if !ok {
			continue
		}
		total := ansi.StringWidth(lines[y])
		from, to = snap(lines[y], from, min(to, total))
		if from >= to {
			continue
		}
		lines[y] = ansi.Cut(lines[y], 0, from) +
			th.ButtonHover(ansi.Strip(ansi.Cut(lines[y], from, to))) +
			ansi.Cut(lines[y], to, total)
	}
	return lines
}
