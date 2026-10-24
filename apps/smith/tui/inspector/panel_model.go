package inspector

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Row is a label and its value.
type Row struct{ Label, Value string }

// Section is a heading and its rows.
type Section struct {
	Heading string
	Rows    []Row
	// Items are plain lines under the heading (a list of tasks or projects).
	Items []string
}

// Action is a clickable line.
type Action struct {
	ID       string
	Label    string
	Shortcut string
	Disabled bool
}

// Panel is what the inspector shows.
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
			parts := pane.Wrap(r.Value, valueW, 0)
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
				text = th.Accent(text)
			}
			hits[len(lines)] = a.ID
			lines = append(lines, text)
		}
	}
	return lines, hits
}
