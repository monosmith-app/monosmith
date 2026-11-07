package selection

import "github.com/monosmith-app/monosmith/apps/smith/tui/layout"

// Cell is one cell of the screen.
type Cell struct{ X, Y int }

// State is a selection, from a press to the release and after it. It is a
// value: every method returns the next State.
type State struct {
	// Anchor is the cell of the press, Head the cell the pointer is on now (kept
	// inside Region).
	Anchor, Head Cell
	// Region is the area the drag may not leave.
	Region layout.Rect
	// Active is true while the button is held.
	Active bool
	// Held is true after the release, while the highlight is kept.
	Held bool
	// moved is true once the pointer left the anchor cell: before it, a click
	// selects nothing.
	moved bool
}

// Press starts a selection at (x, y) in region. Nothing is selected until Drag
// moves the pointer.
func (s State) Press(region layout.Rect, x, y int) State {
	return State{Anchor: Cell{x, y}, Head: Cell{x, y}, Region: region, Active: true}
}

// Drag moves the head to (x, y), kept inside the region. It does nothing when
// no button is held.
func (s State) Drag(x, y int) State {
	if !s.Active {
		return s
	}
	r := s.Region
	s.Head = Cell{clamp(x, r.X, r.X+r.W-1), clamp(y, r.Y, r.Y+r.H-1)}
	s.moved = s.moved || s.Head != s.Anchor
	return s
}

// Release lets go of the button. It reports whether a selection exists; when it
// does not (a plain click) the State is cleared.
func (s State) Release() (State, bool) {
	if !s.Active {
		return s, false
	}
	if !s.moved {
		return State{}, false
	}
	s.Active, s.Held = false, true
	return s, true
}

// Clear drops the selection and its highlight.
func (State) Clear() State { return State{} }

// Selected is true while there are cells to draw or to copy: the drag has
// moved, and the button is still held or the highlight is kept.
func (s State) Selected() bool { return s.moved && (s.Active || s.Held) }

// bounds are the first and the last cell in reading order.
func (s State) bounds() (first, last Cell) {
	first, last = s.Anchor, s.Head
	if last.Y < first.Y || (last.Y == first.Y && last.X < first.X) {
		first, last = last, first
	}
	return first, last
}

// cols are the cells [from, to) of row y that are selected, and false when the
// row is outside the selection. The first row runs from the first cell to the
// right edge of the region, a middle row fills the region, the last row runs
// from the left edge to the last cell (included).
func (s State) cols(y int) (from, to int, ok bool) {
	first, last := s.bounds()
	if !s.Selected() || y < first.Y || y > last.Y {
		return 0, 0, false
	}
	from, to = s.Region.X, s.Region.X+s.Region.W
	if y == first.Y {
		from = first.X
	}
	if y == last.Y {
		to = last.X + 1
	}
	return from, to, from < to
}

func clamp(v, lo, hi int) int { return max(lo, min(v, max(lo, hi))) }
