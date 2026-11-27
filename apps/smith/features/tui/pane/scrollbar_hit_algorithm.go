package pane

// Bar is a scrollbar as the screen shows it: the rows of its column, and the
// part of them the thumb covers.
type Bar struct {
	// Top is the screen row of the first cell of the bar; H is its height.
	Top, H int
	// At is the first cell of the thumb, counted from Top; Size is its height.
	At, Size int
}

// BarAt finds the scrollbar that has the cell at (x, y): a column of track and
// thumb glyphs, as Scrollbar draws them, with at least one of each. cell reads
// one screen cell, and gives "" outside the screen.
//
// Parameters: cell reads the screen; x and y are the cell; track and thumb are
// the theme's glyphs (theme.Icons.Track and Thumb).
// Returns: the bar, and false when the cell is not on one.
func BarAt(cell func(x, y int) string, x, y int, track, thumb string) (Bar, bool) {
	on := func(y int) bool { c := cell(x, y); return c == track || c == thumb }
	if !on(y) {
		return Bar{}, false
	}
	top, end := y, y
	for on(top - 1) {
		top--
	}
	for on(end + 1) {
		end++
	}
	b := Bar{Top: top, H: end - top + 1, At: -1}
	for i := range b.H {
		if cell(x, top+i) == thumb {
			if b.At < 0 {
				b.At = i
			}
			b.Size++
		}
	}
	if b.At < 0 || b.Size == b.H {
		return Bar{}, false
	}
	return b, true
}
