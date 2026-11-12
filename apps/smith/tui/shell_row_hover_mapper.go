package tui

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// rowHoverKinds are the hits of a whole row that a click selects or ticks: a
// project, a task, a recent run, a check, a release item, a dependency and an
// option of a form. Under the pointer they take the one row hover look of the
// theme. Buttons ("action", "choose", "nav") draw their own hover look.
var rowHoverKinds = map[string]bool{
	"project": true, "task": true, "recent": true, "finding": true, "release": true, "dep": true, "option": true,
}

// withRowHover draws the row of c under the pointer in the row hover look, when
// c is drawn in area and the row answers a click. The selected row keeps its
// own look, and nothing behind a dialog looks hovered.
func (s Shell) withRowHover(c content, area layout.Rect) content {
	if s.modalOpen() || s.mouseOff || !area.Contains(s.mouseX, s.mouseY) {
		return c
	}
	y := s.mouseY - area.Y
	h, ok := c.hits[y]
	if !ok || !rowHoverKinds[h.kind] || len(h.spans) > 0 || y >= len(c.lines) || s.selectedHit(h) {
		return c
	}
	c.lines = slices.Clone(c.lines)
	c.lines[y] = s.th.RowHover(pane.Fit(c.lines[y], area.W))
	return c
}

// selectedHit is true when the row of h is the selected one of its list.
func (s Shell) selectedHit(h hit) bool {
	switch h.kind {
	case "project":
		return h.id == s.list.SelectedName()
	case "task":
		return h.id == s.runSel
	case "recent":
		return h.index == s.recentAt
	case "finding":
		return s.doctor.IsCursorRow(h.index)
	case "release":
		return h.index == s.release.Cursor
	case "dep":
		return h.index == s.deps.Cursor
	case "option":
		return h.index == s.guide.Cursor()
	}
	return false
}
