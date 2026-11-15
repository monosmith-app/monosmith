package pointer

import "time"

// DoubleClickWindow is how close in time two presses on the same cell must be.
const DoubleClickWindow = 400 * time.Millisecond

// Click is a press of the left button.
type Click struct {
	X, Y int
	At   time.Time
}

// IsDouble reports whether next is the second press of a double click after prev.
func IsDouble(prev, next Click) bool {
	return !prev.At.IsZero() && prev.X == next.X && prev.Y == next.Y &&
		next.At.Sub(prev.At) <= DoubleClickWindow && next.At.After(prev.At)
}
