package splitter

import "strings"

// Junction is a cell where divisions of the main screen meet, and the glyph it
// is drawn with.
type Junction struct {
	X, Y  int
	Glyph string
}

// Junctions lists the cells where the divisions inside a pane meet: a
// vertical division that starts under the top border of row top ends in a tee
// (┬) in it, one that ends over the bottom border of row bottom in a tee (┴)
// in it, and a horizontal division that starts, ends or crosses a vertical one
// has a tee or a cross there (├, ┤, ┼). The divider beside the detail is the
// border of the list's frame: it meets nothing, so it has no junction.
//
// Parameters: spots are the dividers; top and bottom are the rows of the
// pane's top and bottom borders.
// Returns: the junctions, in the order of the vertical divisions.
func Junctions(spots []Spot, top, bottom int) []Junction {
	var out []Junction
	for _, c := range spots {
		if c.Axis != Column || strings.HasSuffix(c.ID, detailDividerTail) {
			continue
		}
		if c.Hit.Y == top+1 {
			out = append(out, Junction{X: c.Hit.X, Y: top, Glyph: "┬"})
		}
		if c.Hit.Y+c.Hit.H == bottom {
			out = append(out, Junction{X: c.Hit.X, Y: bottom, Glyph: "┴"})
		}
		for _, r := range spots {
			if r.Axis != Row || r.Hit.Y < c.Hit.Y || r.Hit.Y >= c.Hit.Y+c.Hit.H {
				continue
			}
			first, last := r.Hit.X, r.Hit.X+r.Hit.W-1
			switch {
			case c.Hit.X == first:
				out = append(out, Junction{X: c.Hit.X, Y: r.Hit.Y, Glyph: "├"})
			case c.Hit.X == last:
				out = append(out, Junction{X: c.Hit.X, Y: r.Hit.Y, Glyph: "┤"})
			case c.Hit.X > first && c.Hit.X < last:
				out = append(out, Junction{X: c.Hit.X, Y: r.Hit.Y, Glyph: "┼"})
			}
		}
	}
	return out
}
