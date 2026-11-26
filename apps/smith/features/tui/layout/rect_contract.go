package layout

import "github.com/monosmith-app/monosmith/libs/words/words"

// Rect is a rectangle of terminal cells; X and Y are the top-left cell.
type Rect struct{ X, Y, W, H int }

// Contains reports whether the cell (x, y) lies inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return r.W > 0 && r.H > 0 && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Frame is how many cells the frame of a pane takes from each side of it. A
// pane is a rounded box with its title in the top edge: one row at the top and
// at the bottom, and at each side the border and one blank cell of padding.
// Every size and every position of the main screen derives from one Frame.
type Frame struct{ Top, Left, Right, Bottom int }

// Pad is the blank cell between a pane's side border and its content.
const Pad = 1

// MainFrame is the frame of the list, or of the single pane.
var MainFrame = Frame{Top: 1, Left: 1 + Pad, Right: 1 + Pad, Bottom: 1}

// DetailFrame is the frame of the detail beside the list. It is a box of its
// own: the list's right border and the detail's left border stand side by side.
var DetailFrame = Frame{Top: 1, Left: 1 + Pad, Right: 1 + Pad, Bottom: 1}

// Inside is the rectangle of r left after the frame's insets.
func (f Frame) Inside(r Rect) Rect {
	return Rect{X: r.X + f.Left, Y: r.Y + f.Top, W: max(r.W-f.Left-f.Right, 0), H: max(r.H-f.Top-f.Bottom, 0)}
}

// Inner is the rectangle of a pane inside the main frame.
func (r Rect) Inner() Rect { return MainFrame.Inside(r) }

// Area names a part of the screen.
type Area int

// The areas.
const (
	Nowhere Area = iota
	Header
	Tabs
	Main
	Detail
	Footer
	StatusBar
	Rail
)

// String names the area.
func (a Area) String() string {
	return words.Say([...]string{"layout.area.nowhere", "layout.area.header", "layout.area.tabs", "layout.area.main", "layout.area.detail", "layout.area.footer", "layout.area.status", "layout.area.rail"}[a])
}

// Mode is how much of the interface fits.
type Mode int

// Modes by width: Wide shows the rail of views and the list and the detail side
// by side, Narrow the strip of views and only the list (Enter opens the detail
// over it, Esc goes back).
const (
	Wide Mode = iota
	Narrow
)

// WideFrom is the width, in columns, from which the rail of views sits at the
// left and the detail beside the list. 72 is the supported floor (see the shell).
const WideFrom = 100
