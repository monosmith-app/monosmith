package footer

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Placed is a hint and the columns it covers.
type Placed struct {
	Hint commands.Hint
	X, W int
}

// gap is the room between two hints.
const gap = 2

func width(h commands.Hint) int { return pane.Width(h.Glyph) + 1 + pane.Width(h.Label) }

func total(hints []commands.Hint) int {
	n := 0
	for i, h := range hints {
		if i > 0 {
			n += gap
		}
		n += width(h)
	}
	return n
}

// Place lays the hints out in w cells: the view's on the left, the tail on the
// right. When they do not fit the view's last hints go first, then the tail's
// first, so the help and quit keys are the last to go.
func Place(view, tail []commands.Hint, w int) []Placed {
	view, tail = append([]commands.Hint(nil), view...), append([]commands.Hint(nil), tail...)
	fits := func() bool { return 2+total(view)+gap+total(tail) <= w }
	for !fits() && len(view) > 0 {
		view = view[:len(view)-1]
	}
	for !fits() && len(tail) > 0 {
		tail = tail[1:]
	}
	var out []Placed
	x := 1
	for _, h := range view {
		out = append(out, Placed{Hint: h, X: x, W: width(h)})
		x += width(h) + gap
	}
	x = max(w-1-total(tail), x)
	for _, h := range tail {
		out = append(out, Placed{Hint: h, X: x, W: width(h)})
		x += width(h) + gap
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
