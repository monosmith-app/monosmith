package resultdialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Title is the frame title of the dialog.
func (m Model) Title() string { return m.title }

// Lines draws the dialog, framed, at the size Place set.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 24 || r.H < 8 {
		return nil
	}
	inner := r.W - 4
	var body []string
	for _, f := range m.facts {
		text := f.Text
		switch f.Kind {
		case Good:
			text = th.Success(th.Icons.Passed) + " " + text
		case Bad:
			text = th.Warning(th.Icons.Failed+" ") + th.Bold(text)
		default:
			text = th.Muted(text)
		}
		body = append(body, pane.Wrap(text, inner, 2)...)
	}
	body = append(body, "")
	for i, c := range m.choices {
		mark, label := "  ", c.label
		if i == m.cursor {
			mark, label = th.Accent(th.Icons.Pointer+" "), th.Bold(label)
		}
		body = append(body, pane.Wrap(mark+label+"  "+th.Muted(c.hint), inner, 4)...)
	}
	body = append(body, "", th.Muted(m.foot), th.Muted("up/down choose · enter select · esc closes and changes nothing"))
	rows := make([]string, 0, r.H-2)
	for _, l := range body {
		if len(rows) >= r.H-2 {
			break
		}
		rows = append(rows, " "+pane.Fit(l, inner))
	}
	return pane.Frame(th, m.title, rows, r.W, r.H, true)
}
