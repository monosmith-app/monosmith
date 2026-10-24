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

// The areas, in the order focus moves through them.
const (
	Nowhere Area = iota
	Footer
	Navigator
	Main
	Inspector
	StatusBar
)

// String names the area.
func (a Area) String() string {
	return [...]string{"nowhere", "footer", "navigator", "main", "inspector", "status bar"}[a]
}

// Mode is how much of the interface fits.
type Mode int

// Modes by width: Wide has the navigator and the inspector, Medium folds the
// inspector into the main view, Narrow also turns the navigator into a tab.
const (
	Wide Mode = iota
	Medium
	Narrow
)

// Breakpoints in columns: the inspector folds below WideFrom, the navigator
// becomes a tab below MediumFrom. 80 is the supported floor.
const (
	WideFrom   = 100
	MediumFrom = 80
)
