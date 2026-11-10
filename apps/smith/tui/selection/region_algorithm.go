package selection

import "github.com/monosmith-app/monosmith/apps/smith/tui/layout"

// Cut is a divider: the column (Vertical) or the row of the cell (X, Y). The
// divider cell itself belongs to no region.
type Cut struct {
	X, Y     int
	Vertical bool
}

// Divide cuts each area along the dividers that lie inside it, so that the
// panels a divider separates are regions of their own. The columns are cut
// first: a row divider may span only one side of a column.
func Divide(areas []layout.Rect, cuts []Cut) []layout.Rect {
	for _, vertical := range []bool{true, false} {
		for _, c := range cuts {
			if c.Vertical != vertical {
				continue
			}
			var next []layout.Rect
			for _, r := range areas {
				next = append(next, split(r, c)...)
			}
			areas = next
		}
	}
	return areas
}

// split cuts r at c; an area the cut does not touch comes back whole, and an
// empty half is dropped.
func split(r layout.Rect, c Cut) []layout.Rect {
	if !r.Contains(c.X, c.Y) {
		return []layout.Rect{r}
	}
	var halves []layout.Rect
	if c.Vertical {
		halves = []layout.Rect{
			{X: r.X, Y: r.Y, W: c.X - r.X, H: r.H},
			{X: c.X + 1, Y: r.Y, W: r.X + r.W - c.X - 1, H: r.H},
		}
	} else {
		halves = []layout.Rect{
			{X: r.X, Y: r.Y, W: r.W, H: c.Y - r.Y},
			{X: r.X, Y: c.Y + 1, W: r.W, H: r.Y + r.H - c.Y - 1},
		}
	}
	out := halves[:0]
	for _, h := range halves {
		if h.W > 0 && h.H > 0 {
			out = append(out, h)
		}
	}
	return out
}
