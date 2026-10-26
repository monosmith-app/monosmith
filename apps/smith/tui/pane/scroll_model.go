package pane

// Scroll is where a list's window starts, kept apart from the list's
// selection so the wheel can move the one without the other. It is a pointer
// on purpose: a list draws from a copy of its model, and the offset must
// survive between frames.
//
// The window follows the selection only when the selection changed since the
// last frame, and then only as far as it must (ScrollTo): a click on a row
// in view does not move the list, a key that leaves the view scrolls it just
// enough, and the wheel never touches the selection.
type Scroll struct {
	top, seen int
	seenSet   bool
}

// Show returns the first row to draw for a selection sel in a window of h rows
// over n rows, and remembers it. Call it once per frame, while drawing.
func (s *Scroll) Show(sel, h, n int) int {
	if s == nil {
		return ScrollTo(0, sel, h, n)
	}
	if !s.seenSet || sel != s.seen {
		s.top = ScrollTo(s.top, sel, h, n)
		s.seen, s.seenSet = sel, true
	}
	s.top = min(max(s.top, 0), max(n-h, 0))
	return s.top
}

// By moves the window by delta rows, whatever the selection; the next frame
// keeps it within the rows there are.
func (s *Scroll) By(delta int) {
	if s != nil {
		s.top = max(s.top+delta, 0)
	}
}
