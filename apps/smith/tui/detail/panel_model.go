package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Row is a label and its value.
type Row struct{ Label, Value string }

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

// Section is a heading and its rows.
type Section struct {
	Heading string
	Rows    []Row
	// Items are plain lines under the heading (a list of tasks or projects).
	Items []string
	// OneLine keeps each item on one line, ending in an ellipsis when it is
	// too long, instead of wrapping it (a command is not prose).
	OneLine bool
	// Bullets puts a "•" before each item and wraps it under its text.
	Bullets bool
	// Collapse is how many items show until the panel is Expanded; the rest
	// is told as a count. Zero shows them all.
	Collapse int
}

// Action is a clickable line.
type Action struct {
	ID       string
	Label    string
	Shortcut string
	Disabled bool
}

// Panel is what the detail shows.
type Panel struct {
	Sections []Section
	Actions  []Action
	// After are the sections drawn under the actions: what is long and
	// secondary, so it can never push the actions off the screen.
	After []Section
	// Expanded shows every item of a Collapse section.
	Expanded bool
}

// View draws the panel in w cells and returns it with the row of each action
// (relative to the first line), so a click can be mapped to it. hover is the
// ID of the action under the pointer. The lines are all of the panel: the
// caller shows a window of them (see Window).
func (p Panel) View(th theme.Theme, w int, hover string) (lines []string, hits map[int]string) {
	hits = map[int]string{}
	if len(p.Sections) == 0 && len(p.Actions) == 0 && len(p.After) == 0 {
		return []string{th.Muted(" Nothing selected.")}, hits
	}
	labelW := 0
	for _, list := range [][]Section{p.Sections, p.After} {
		for _, s := range list {
			for _, r := range s.Rows {
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
		lines = append(lines, " "+th.Title("Actions"))
		for _, a := range p.Actions {
			text := a.Label
			if a.Shortcut != "" {
				gap := max(w-3-pane.Width(a.Label)-pane.Width(a.Shortcut), 1)
				text += strings.Repeat(" ", gap) + a.Shortcut
			}
			text = pane.Fit(" "+text, w)
			switch {
			case a.Disabled:
				text = th.Muted(text)
			case a.ID == hover:
				text = th.Selected(text)
			default:
				text = th.Bold(text)
			}
			hits[len(lines)] = a.ID
			lines = append(lines, text)
		}
	}
	for _, s := range p.After {
		lines = p.section(lines, th, s, w, labelW)
	}
	return lines, hits
}

// section appends one section's lines.
func (p Panel) section(lines []string, th theme.Theme, s Section, w, labelW int) []string {
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	folded := s.Collapse > 0 && !p.Expanded && len(s.Items) > s.Collapse
	if s.Heading != "" {
		heading := " " + th.Title(s.Heading)
		if folded {
			// Said at the top too: the count under the list can be below the fold.
			heading += th.Muted(fmt.Sprintf("  %d, e shows all", len(s.Items)))
		}
		lines = append(lines, heading)
	}
	for _, r := range s.Rows {
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
	switch {
	case hidden > 0:
		lines = append(lines, pane.Fit(th.Muted(fmt.Sprintf("  +%d more (e shows all)", hidden)), w))
	case s.Collapse > 0 && p.Expanded && len(items) > s.Collapse:
		lines = append(lines, pane.Fit(th.Muted("  (e shows fewer)"), w))
	}
	return lines
}
