package splitter

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"maps"
)

// State is what the person did to the dividers: the sizes they set, the divider
// the keys move and the one the pointer holds. It is a value; every method
// returns a new State and leaves the maps of the old one alone.
type State struct {
	// Sizes is the size each divider was given, by ID. A divider with no entry is
	// automatic.
	Sizes map[string]int
	// Selected is the divider the keys move ("" for none); Dragging is the one
	// the pointer holds ("" for none).
	Selected, Dragging string
}

// New is a State that starts from sizes kept earlier (nil for none).
func New(sizes map[string]int) State { return State{Sizes: maps.Clone(sizes)} }

// Size is the size given to a divider, and whether one was given.
func (s State) Size(id string) (int, bool) {
	n, ok := s.Sizes[id]
	return n, ok
}

// set records a size for a divider.
func (s State) set(p Spot, size int) State {
	sizes := maps.Clone(s.Sizes)
	if sizes == nil {
		sizes = map[string]int{}
	}
	sizes[p.ID] = pane.Clamp(size, p.Min, p.Max)
	s.Sizes = sizes
	return s
}

// find is the spot with an ID.
func find(spots []Spot, id string) (Spot, bool) {
	for _, p := range spots {
		if p.ID == id {
			return p, true
		}
	}
	return Spot{}, false
}

// SelectedIn is true while the selected divider is one of spots.
func (s State) SelectedIn(spots []Spot) bool {
	_, ok := find(spots, s.Selected)
	return ok
}

// Look is how the divider id is drawn: held, selected, or at rest.
func (s State) Look(id string) Look {
	switch id {
	case "":
		return Resting
	case s.Dragging:
		return Dragging
	case s.Selected:
		return Selected
	}
	return Resting
}

// Select moves the selection to the next (dir 1) or the previous (dir -1) divider
// of an axis, wrapping round. With none selected it takes the first (or the last,
// for dir -1). It reports false, and changes nothing, when the screen has no
// divider of that axis.
func (s State) Select(spots []Spot, axis Axis, dir int) (State, bool) {
	var ids []string
	for _, p := range spots {
		if p.Axis == axis {
			ids = append(ids, p.ID)
		}
	}
	if len(ids) == 0 {
		return s, false
	}
	at := -1
	for i, id := range ids {
		if id == s.Selected {
			at = i
		}
	}
	switch {
	case at < 0 && dir < 0:
		at = len(ids) - 1
	case at < 0:
		at = 0
	default:
		at = ((at+dir)%len(ids) + len(ids)) % len(ids)
	}
	s.Selected = ids[at]
	return s, true
}

// Cycle selects the next (dir 1) or the previous (dir -1) divider of any axis,
// wrapping round. With none selected it takes the first (or the last, for dir
// -1). It reports false when the screen has no divider.
func (s State) Cycle(spots []Spot, dir int) (State, bool) {
	if len(spots) == 0 {
		return s, false
	}
	at := -1
	for i, p := range spots {
		if p.ID == s.Selected {
			at = i
		}
	}
	if at < 0 && dir < 0 {
		at = 0
	}
	s.Selected = spots[((at+dir)%len(spots)+len(spots))%len(spots)].ID
	return s, true
}

// Deselect lets go of the selection.
func (s State) Deselect() State { s.Selected = ""; return s }

// Move moves the selected divider by step cells towards the right or the
// bottom (dir 1) or the left or the top (dir -1). With nothing selected it moves
// the only divider, or else the first one and selects it. It returns the spot as
// it was before the move and false when the screen has no divider.
func (s State) Move(spots []Spot, dir, step int) (State, Spot, bool) {
	p, ok := find(spots, s.Selected)
	if !ok {
		if len(spots) == 0 {
			return s, Spot{}, false
		}
		p = spots[0]
		s.Selected = p.ID
	}
	return s.set(p, p.Size+p.Sign*dir*step), p, true
}

// Press grabs the divider under (x, y), if any, and selects it.
func (s State) Press(spots []Spot, x, y int) (State, bool) {
	for _, p := range spots {
		if p.Hit.Contains(x, y) {
			s.Dragging, s.Selected = p.ID, p.ID
			return s, true
		}
	}
	return s, false
}

// Drag moves the held divider to the pointer. Nothing held changes nothing.
func (s State) Drag(spots []Spot, x, y int) State {
	if p, ok := find(spots, s.Dragging); ok {
		return s.set(p, p.SizeAt(x, y))
	}
	return s
}

// Release lets go of the divider and says which one it was ("" when none was held),
// so the caller can keep the size.
func (s State) Release() (State, string) {
	id := s.Dragging
	s.Dragging = ""
	return s, id
}
