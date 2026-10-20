package dialog

// Outcome is how a dialog ended.
type Outcome struct {
	// Done is true once the dialog should close.
	Done bool
	// Confirmed is true when the user chose to go ahead.
	Confirmed bool
}

// Rect is where a dialog sits on screen.
type Rect struct{ X, Y, W, H int }

// Contains reports whether the cell is inside.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}
