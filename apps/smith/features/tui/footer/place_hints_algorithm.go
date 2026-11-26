package footer

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

// Placed is a hint and the columns it covers.
type Placed struct {
	Hint commands.Hint
	X, W int
}

// gap is the room between two hints: a space, the separator and a space.
const gap = 3

// keptLast is how many hints at the end of the tail (help and quit) stay
// after the view's own hints are gone.
const keptLast = 2

// width is the cells a hint takes: its key, a space and its description, in
// every theme.
func width(_ theme.Theme, h commands.Hint) int {
	return pane.Width(h.Glyph) + 1 + pane.Width(h.Label)
}

func total(th theme.Theme, hints []commands.Hint) int {
	n := 0
	for i, h := range hints {
		if i > 0 {
			n += gap
		}
		n += width(th, h)
	}
	return n
}

// Place lays the hints out in w cells as one run, each as wide as it is drawn
// in th: the view's first, then the tail ("key desc • … • ? help • q quit").
// When they do not fit the lowest-priority hints go first: the tail's global
// hints, from its left (views, mouse, palette), so the view keeps its own keys
// (Home keeps next (Enter)); then the view's last hints go; the last keptLast
// hints of the tail (help and quit) are the last of all to go.
//
// Parameters: th is the theme; view are the open view's hints, most important
// first; tail are the global hints; w is the width of the row.
// Returns: the hints that fit, with their columns.
func Place(th theme.Theme, view, tail []commands.Hint, w int) []Placed {
	view, tail = append([]commands.Hint(nil), view...), append([]commands.Hint(nil), tail...)
	fits := func() bool { return 2+total(th, view)+gap+total(th, tail) <= w }
	for !fits() && len(tail) > keptLast {
		tail = tail[1:]
	}
	for !fits() && len(view) > 0 {
		view = view[:len(view)-1]
	}
	for !fits() && len(tail) > 0 {
		tail = tail[1:]
	}
	var out []Placed
	x := 1
	for _, h := range view {
		out = append(out, Placed{Hint: h, X: x, W: width(th, h)})
		x += width(th, h) + gap
	}
	for _, h := range tail {
		out = append(out, Placed{Hint: h, X: x, W: width(th, h)})
		x += width(th, h) + gap
	}
	return out
}

// At is the hint under column x, if any.
func At(placed []Placed, x int) (commands.Hint, bool) {
	for _, p := range placed {
		if x >= p.X && x < p.X+p.W {
			return p.Hint, true
		}
	}
	return commands.Hint{}, false
}
