package splitter

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Axis is the direction a divider runs.
type Axis int

const (
	// Column is a vertical divider between two panels side by side; it moves
	// left and right.
	Column Axis = iota
	// Row is a horizontal divider between two stacked panels; it moves up and
	// down.
	Row
)

// Spot is one divider on the screen as drawn now. A screen builds its spots from
// its layout and hands them to State; the component never reads the layout.
type Spot struct {
	// ID names the divider and is the key its size is kept under.
	ID   string
	Axis Axis
	// Hit is the cells a press grabs the divider by.
	Hit layout.Rect
	// Size is the cells of the panel the divider sizes, Min and Max what it may
	// be. A screen that has no room for the range gives Min == Max.
	Size, Min, Max int
	// Origin and Sign turn a pointer position p (a column for Column, a row for
	// Row) into that size: Sign * (p - Origin). Sign is +1 when the panel is
	// before the divider and -1 when it is after it.
	Origin, Sign int
}

// at is the pointer coordinate that counts for the divider's axis.
func (p Spot) at(x, y int) int {
	if p.Axis == Row {
		return y
	}
	return x
}

// SizeAt is the size the panel gets when the divider is at (x, y), kept
// between Min and Max.
func (p Spot) SizeAt(x, y int) int { return pane.Clamp(p.Sign*(p.at(x, y)-p.Origin), p.Min, p.Max) }

// Look is how a divider is drawn.
type Look int

const (
	// Resting is a divider nothing is happening to; it keeps the pane's own rule.
	Resting Look = iota
	// Selected is the divider the keys move.
	Selected
	// Dragging is the divider held by the pointer.
	Dragging
)
