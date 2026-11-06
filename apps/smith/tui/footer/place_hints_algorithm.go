package footer

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Placed is a hint and the columns it covers.
type Placed struct {
	Hint commands.Hint
	X, W int
}

// gap is the room between two hints.
const gap = 2

// width is the cells a hint takes: a button when a click presses its key, the
// glyph and the label as text when not.
func width(th theme.Theme, h commands.Hint) int {
	if h.Key != "" {
		return button.Width(th, h.Label, h.Glyph)
	}
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

// Place lays the hints out in w cells, each as wide as it is drawn in th: the view's on the left, the tail on the
// right. When they do not fit the view's last hints go first, then the tail's
// first, so the help and quit keys are the last to go.
func Place(th theme.Theme, view, tail []commands.Hint, w int) []Placed {
	view, tail = append([]commands.Hint(nil), view...), append([]commands.Hint(nil), tail...)
	fits := func() bool { return 2+total(th, view)+gap+total(th, tail) <= w }
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
	x = max(w-1-total(th, tail), x)
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
