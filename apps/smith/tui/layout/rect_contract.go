package layout

// Rect is a rectangle of terminal cells; X and Y are the top-left cell.
type Rect struct{ X, Y, W, H int }

// Contains reports whether the cell (x, y) lies inside the rectangle.
func (r Rect) Contains(x, y int) bool {
	return r.W > 0 && r.H > 0 && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Inner is the rectangle inside a one-cell border.
func (r Rect) Inner() Rect {
	return Rect{X: r.X + 1, Y: r.Y + 1, W: max(r.W-2, 0), H: max(r.H-2, 0)}
}

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
)

// String names the area.
func (a Area) String() string {
	return [...]string{"nowhere", "header", "tabs", "main", "detail", "footer", "status bar"}[a]
}

// Mode is how much of the interface fits.
type Mode int

// Modes by width: Wide shows the list and the detail side by side, Narrow only
// the list (Enter opens the detail over it, Esc goes back).
const (
	Wide Mode = iota
	Narrow
)

// WideFrom is the width, in columns, from which the detail sits beside the
// list. 72 is the supported floor (see the shell).
const WideFrom = 100

// TallFrom is the height, in rows, from which the tabs are drawn as rounded
// boxes three rows high; below it they are one row.
const TallFrom = 24
