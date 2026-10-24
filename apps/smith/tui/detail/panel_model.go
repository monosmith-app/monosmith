package detail

import (
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
}

// View draws the panel in w cells and returns it with the row of each action
// (relative to the first line), so a click can be mapped to it. hover is the
// ID of the action under the pointer.
func (p Panel) View(th theme.Theme, w int, hover string) (lines []string, hits map[int]string) {
	hits = map[int]string{}
	if len(p.Sections) == 0 && len(p.Actions) == 0 {
		return []string{th.Muted(" Nothing selected.")}, hits
	}
	labelW := 0
	for _, s := range p.Sections {
		for _, r := range s.Rows {
			labelW = max(labelW, pane.Width(r.Label))
		}
	}
	labelW = min(labelW, max(w/2-1, 4))
	for i, s := range p.Sections {
		if i > 0 {
			lines = append(lines, "")
		}
		if s.Heading != "" {
			lines = append(lines, " "+th.Title(s.Heading))
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
		for _, it := range s.Items {
			if s.OneLine {
				lines = append(lines, pane.Fit(ansi.Truncate("  "+it, w, "…"), w))
				continue
			}
			for _, part := range pane.Wrap("  "+it, w, 2) {
				lines = append(lines, pane.Fit(part, w))
			}
		}
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
	return lines, hits
}
