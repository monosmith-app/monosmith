package splitter

// Junction is a cell where divisions of the main screen meet, and the glyph it
// is drawn with.
type Junction struct {
	X, Y  int
	Glyph string
}

// Junctions lists the cells where the divisions meet: a vertical division that
// starts under the rule of row ruleY ends in a tee (┬) in the rule, and a
// horizontal division that starts, ends or crosses a vertical one has a tee or
// a cross there (├, ┤, ┼). With no bottom edge, no division ends in ┴.
func Junctions(spots []Spot, ruleY int) []Junction {
	var out []Junction
	for _, c := range spots {
		if c.Axis != Column {
			continue
		}
		if c.Hit.Y == ruleY+1 {
			out = append(out, Junction{X: c.Hit.X, Y: ruleY, Glyph: "┬"})
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
