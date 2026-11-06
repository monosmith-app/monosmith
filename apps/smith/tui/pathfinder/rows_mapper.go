package pathfinder

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Hint is the line of keys to show under the list.
var Hint = words.Say("pathfinder.hint")

// DefaultRows is how many rows of the list are drawn before the rest is
// folded into a "more" line.
const DefaultRows = 8

// Line is one drawn line of the list. Index is the row it stands for, or -1
// for a line that is only text (the place, the "more" note).
type Line struct {
	Text  string
	Index int
}

// Rows draws the list in w cells and at most limit rows. The window follows
// the highlight, and what does not fit is counted in a last line, so even a
// folder of thousands of names is a short list.
//
// Parameters: th is the theme; w is the width in cells; limit is how many rows
// to draw (DefaultRows when 0 or less).
// Returns: the lines to show, the first one naming the folder.
func (m Model) Rows(th theme.Theme, w, limit int) []Line {
	if limit <= 0 {
		limit = DefaultRows
	}
	where := words.Say("pathfinder.here")
	if m.dir != "" {
		where = m.dir
	}
	out := []Line{{Text: pane.Fit(" "+th.Muted(words.Say("pathfinder.in", "where", where)), w), Index: -1}}
	switch {
	case m.unreadable:
		return append(out, Line{Text: pane.Fit("  "+th.Muted(words.Say("pathfinder.unreadable")), w), Index: -1})
	case len(m.rows) == 0:
		return append(out, Line{Text: pane.Fit("  "+th.Muted(words.Say("pathfinder.nomatch")), w), Index: -1})
	}
	top := max(m.cursor-limit+1, 0)
	end := min(top+limit, len(m.rows))
	for i := top; i < end; i++ {
		out = append(out, Line{Text: pane.Fit(m.rowText(th, i), w), Index: i})
	}
	if hidden := len(m.rows) - (end - top); hidden > 0 {
		note := words.Say("pathfinder.more", "count", hidden)
		if m.truncated {
			note = words.Say("pathfinder.more.truncated", "count", hidden, "max", maxNames)
		}
		out = append(out, Line{Text: pane.Fit("  "+th.Muted(words.Say("pathfinder.note", "note", note)), w), Index: -1})
	}
	return out
}

// rowText draws row i: the pointer, the name (a folder ends in a slash, a
// git-ignored one is dimmed) and a dimmed hint.
func (m Model) rowText(th theme.Theme, i int) string {
	r := m.rows[i]
	name := r.name
	if r.kind == rowFolder {
		name += "/"
	}
	pointer := "  "
	switch {
	case i == m.cursor:
		pointer, name = th.Accent(th.Icons.Pointer+" "), th.Accent(name)
	case r.ignored:
		name = th.Faint(name)
	case r.kind == rowFolder:
		name = th.Info(name)
	case r.kind == rowSetting:
		name = th.Bold(name)
	}
	if r.hint != "" {
		name += "  " + th.Muted(r.hint)
	}
	return pointer + name
}
