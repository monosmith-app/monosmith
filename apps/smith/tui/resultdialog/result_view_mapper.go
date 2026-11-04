package resultdialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Title is the frame title of the dialog.
func (m Model) Title() string { return m.title }

// Hint is the line under the choices that says how to answer.
const Hint = "up/down choose · enter select · click an entry, double-click to select · esc closes and changes nothing"

// Lines draws the dialog, framed, at the size Place set.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 24 || r.H < 8 {
		return nil
	}
	body, _ := m.body(th)
	rows := make([]string, 0, r.H-2)
	for _, l := range body {
		if len(rows) >= r.H-2 {
			break
		}
		rows = append(rows, " "+pane.Fit(l, r.W-4))
	}
	return pane.Frame(th, m.title, rows, r.W, r.H, true)
}

// span is the lines a choice takes in the body, first included, last excluded.
type span struct{ from, to int }

// body is the wrapped text lines of the dialog and, for each choice, the lines
// it takes: Lines draws the first and Click maps a row to a choice with the
// second, so what is drawn and what is clickable never differ.
func (m Model) body(th theme.Theme) (body []string, spans []span) {
	inner := m.rect.W - 4
	if w := m.workingLine(th); w != "" {
		body = append(body, pane.Wrap(w, inner, 2)...)
	}
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
		from := len(body)
		body = append(body, pane.Wrap(mark+label+"  "+th.Muted(c.hint), inner, 4)...)
		spans = append(spans, span{from, len(body)})
	}
	body = append(body, "", th.Muted(m.foot))
	body = append(body, pane.Wrap(th.Muted(Hint), inner, 0)...)
	return body, spans
}

// Click handles a left press: a press on a choice moves the cursor to it, a
// second press on the same cell (double is true) selects it like Enter. A press
// elsewhere changes nothing.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen
// cells; double says the press is the second of a double click.
// Returns: the new dialog and what the host must do.
func (m Model) Click(th theme.Theme, x, y int, double bool) (Model, Outcome) {
	if !m.rect.Contains(x, y) {
		return m, Outcome{}
	}
	_, spans := m.body(th)
	row := y - m.rect.Y - 1
	for i, sp := range spans {
		if row < m.rect.H-2 && row >= sp.from && row < sp.to {
			m.cursor = i
			if double {
				return m.Key("enter")
			}
			return m, Outcome{}
		}
	}
	return m, Outcome{}
}
