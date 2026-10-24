package commands

import "slices"

// Hint is one entry of the footer: the key as drawn, what it does, and the
// key a click on it presses ("" when the glyph stands for several keys).
type Hint struct {
	Glyph, Label, Key string
}

// Hints are the footer's hints for the open view: first the keys of the view,
// then the Tail, which every view ends with. Only what can be done now is
// listed, so no hint promises a key that would be refused.
func Hints(s State) (view, tail []Hint) {
	for _, c := range all {
		if c.Footer == "" || c.Unavailable(s) != "" || (c.Kind != PaneKey && !c.Visible(s)) {
			continue
		}
		show := c.ShowIn
		if len(show) == 0 {
			show = c.Views
		}
		switch {
		case c.Tail:
			tail = append(tail, hintOf(c))
		case slices.Contains(show, s.View):
			view = append(view, hintOf(c))
		}
	}
	return view, tail
}

func hintOf(c Command) Hint {
	h := Hint{Glyph: c.Glyph, Label: c.Footer}
	// A glyph that is no key name ("1-9", "↑↓") stands for several keys.
	if h.Glyph == "" || h.Glyph == KeyLabel(c.Keys[0]) {
		h.Glyph, h.Key = KeyLabel(c.Keys[0]), c.Keys[0]
	}
	return h
}
