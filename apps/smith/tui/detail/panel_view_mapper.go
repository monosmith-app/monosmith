package detail

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// valueLines breaks a value into lines of valueW cells. A value of several
// words wraps at the spaces; a single long word (a path, a package name) is
// shortened in the middle instead of being cut into fragments.
func valueLines(value string, valueW int) []string {
	if pane.Width(value) <= valueW {
		return []string{value}
	}
	if !strings.ContainsAny(pane.Strip(value), " \t") {
		return []string{pane.MiddleTruncate(value, valueW)}
	}
	lines := pane.Wrap(value, valueW, 0)
	for i, l := range lines {
		if pane.Width(l) > valueW {
			lines[i] = pane.MiddleTruncate(l, valueW)
		}
	}
	return lines
}

// View draws the panel in w cells and returns it with the row of each action
// (relative to the first line), so a click can be mapped to it. hover is the
// ID of the action under the pointer. The lines are all of the panel: the
// caller shows a window of them (see Window).
func (p Panel) View(th theme.Theme, w int, hover string) (lines []string, hits map[int]string) {
	hits = map[int]string{}
	if len(p.Sections) == 0 && len(p.Actions) == 0 && len(p.After) == 0 {
		return []string{th.Muted(words.Say("detail.nothing"))}, hits
	}
	labelW := 0
	for _, list := range [][]Section{p.Sections, p.After} {
		for _, s := range list {
			for _, r := range p.rows(s) {
				labelW = max(labelW, pane.Width(r.Label))
			}
		}
	}
	labelW = min(labelW, max(w/2-1, 4))
	for _, s := range p.Sections {
		lines = p.section(lines, th, s, w, labelW)
	}
	if len(p.Actions) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		title := p.ActionsTitle
		if title == "" {
			title = words.Say("detail.actions")
		}
		lines = append(lines, " "+th.Title(title))
		for _, a := range p.Actions {
			lines = p.action(lines, hits, th, a, w, hover)
		}
	}
	for _, s := range p.After {
		lines = p.section(lines, th, s, w, labelW)
	}
	return lines, hits
}

// action appends the lines of one action: its button, then what it does in
// muted text under it. Every line clicks like the action. The button is the
// shared one (see package button) in the state the action is in: Disabled, Active
// while it is pressed under the pointer, Hover under the pointer or at the
// keyboard cursor, else Rest. With a Cursor the rows lead with the cursor mark.
func (p Panel) action(lines []string, hits map[int]string, th theme.Theme, a Action, w int, hover string) []string {
	mark := " "
	if p.Cursor != "" {
		mark = choicerow.Cursor(th, a.ID == p.Cursor)
	}
	markW := pane.Width(mark)
	label := a.Label
	var notes []string
	if over := button.Width(th, label, a.Shortcut) - max(w-markW, 4); over > 0 {
		// The button is one row: a label too long for it is cut, and said whole under it.
		label = ansi.Truncate(label, max(pane.Width(label)-over, 1), "…")
		notes = append(notes, a.Label)
	}
	if a.Hint != "" {
		notes = append(notes, a.Hint)
	}
	state := button.Rest
	switch {
	case a.Disabled:
		state = button.Disabled
	case a.ID == hover && a.ID == p.Pressed:
		state = button.Active
	case a.ID == hover || a.ID == p.Cursor:
		state = button.Hover
	}
	hits[len(lines)] = a.ID
	lines = append(lines, pane.Fit(mark+button.Render(th, label, a.Shortcut, state), w))
	for _, note := range notes {
		for _, part := range pane.Wrap(note, max(w-markW-4, 4), 0) {
			hits[len(lines)] = a.ID
			lines = append(lines, pane.Fit(strings.Repeat(" ", markW+2)+th.Muted(part), w))
		}
	}
	return lines
}

// rows are the rows of a section the panel draws: all of them, except the
// ones the list beside the panel already shows.
func (p Panel) rows(s Section) []Row {
	if !p.ListVisible {
		return s.Rows
	}
	var out []Row
	for _, r := range s.Rows {
		if !r.ShownByList {
			out = append(out, r)
		}
	}
	return out
}

// section appends one section's lines.
func (p Panel) section(lines []string, th theme.Theme, s Section, w, labelW int) []string {
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	folded := s.Collapse > 0 && !p.Expanded && len(s.Items) > s.Collapse
	foldedAway := s.Fold && !p.Expanded
	if s.Heading != "" {
		// A long heading (a doctor check's name is a sentence) wraps instead of being cut.
		parts := pane.Wrap(s.Heading, max(w-1, 4), 0)
		for _, part := range parts {
			lines = append(lines, " "+th.Title(part))
		}
		if folded || foldedAway {
			// Said at the top too: the count under the list can be below the fold.
			count := words.Say("detail.folded", "count", len(s.Items))
			if foldedAway {
				count = words.Say("detail.fold")
			}
			last := len(lines) - 1
			if pane.Width(parts[len(parts)-1])+pane.Width(count) < w {
				lines[last] += th.Muted(count)
			} else {
				lines = append(lines, " "+th.Muted(strings.TrimLeft(count, " ")))
			}
		}
	}
	if foldedAway {
		return lines
	}
	for _, r := range p.rows(s) {
		// A long value wraps under itself instead of running off the pane.
		label := pane.Fit(th.Muted(r.Label), labelW)
		valueW := max(w-labelW-2, 4)
		parts := valueLines(r.Value, valueW)
		if len(parts) == 0 {
			parts = []string{""}
		}
		for n, part := range parts {
			if n == 0 {
				lines = append(lines, pane.Fit(" "+label+" "+part, w))
				continue
			}
			lines = append(lines, pane.Fit(strings.Repeat(" ", labelW+2)+part, w))
		}
	}
	items := s.Items
	hidden := 0
	if s.Collapse > 0 && !p.Expanded && len(items) > s.Collapse {
		hidden = len(items) - s.Collapse
		items = items[:s.Collapse]
	}
	for _, it := range items {
		switch {
		case s.OneLine:
			lines = append(lines, pane.Fit(ansi.Truncate("  "+it, w, "…"), w))
		case s.Bullets:
			for _, part := range pane.Wrap("  • "+it, w, 4) {
				lines = append(lines, pane.Fit(part, w))
			}
		default:
			for _, part := range pane.Wrap("  "+it, w, 2) {
				lines = append(lines, pane.Fit(part, w))
			}
		}
	}
	for _, it := range s.Tail {
		for _, part := range pane.Wrap("  "+it, w, 2) {
			lines = append(lines, pane.Fit(part, w))
		}
	}
	switch {
	case hidden > 0:
		lines = append(lines, pane.Fit(th.Muted(words.Say("detail.hidden", "count", hidden)), w))
	case s.Collapse > 0 && p.Expanded && len(items) > s.Collapse:
		lines = append(lines, pane.Fit(th.Muted(words.Say("detail.fewer")), w))
	}
	return lines
}
