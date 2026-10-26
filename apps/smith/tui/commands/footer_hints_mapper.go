package commands

import (
	"fmt"
	"slices"
)

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
			tail = append(tail, hintOf(c, s))
		case slices.Contains(show, s.View):
			view = append(view, hintOf(c, s))
		}
	}
	return view, tail
}

func hintOf(c Command, s State) Hint {
	h := Hint{Glyph: c.Glyph, Label: scopedLabel(c, s)}
	// A glyph that is no key name ("1-9", "↑↓") stands for several keys.
	if h.Glyph == "" || h.Glyph == KeyLabel(c.Keys[0]) {
		h.Glyph, h.Key = KeyLabel(c.Keys[0]), c.Keys[0]
	}
	return h
}

// maxScopeName is the longest project name the footer spells out.
const maxScopeName = 16

// scopedLabel is the footer label of a command, with the scope of a run in it:
// "build web" runs on web, "build all 23" on every project, so the scope is
// known before the key is pressed.
func scopedLabel(c Command, s State) string {
	if c.Kind != RunTask {
		return c.Footer
	}
	switch {
	case c.All && s.Projects > 0:
		return fmt.Sprintf("%s %d", c.Footer, s.Projects)
	case !c.All && c.ScopeInFooter && s.Project != "":
		return c.Footer + " " + clip(s.Project, maxScopeName)
	}
	return c.Footer
}

// clip shortens a name to n cells with an ellipsis.
func clip(name string, n int) string {
	r := []rune(name)
	if len(r) <= n {
		return name
	}
	return string(r[:n-1]) + "…"
}
