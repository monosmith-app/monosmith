package resultdialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Title is the frame title of the dialog.
func (m Model) Title() string { return m.title }

// Hint is the line under the choices that says how to answer.
var Hint = words.Say("resultdialog.hint")

// Lines draws the dialog, framed, at the size Place set. A body taller than
// the dialog scrolls under the shared scrollbar; the hint stays on the last rows.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 24 || r.H < 4 {
		return nil
	}
	body, _ := m.body(th)
	rows := pane.DialogBody(th, body, m.hintLines(th), pane.DialogRowWidth(r.W), r.H-2, m.top)
	for i, l := range rows {
		rows[i] = " " + pane.Fit(l, pane.DialogRowWidth(r.W))
	}
	return pane.Frame(th, m.title, rows, r.W, r.H, true)
}

// hintLines are the lines of the hint, pinned under the body.
func (m Model) hintLines(th theme.Theme) []string {
	return pane.Wrap(th.Muted(Hint), max(m.rect.W-4, 1), 0)
}

// room is the rows the scrolling body gets.
func (m Model) room() int {
	return pane.BodyRoom(m.rect.H-2, len(m.hintLines(theme.Default())))
}

// follow scrolls the body so every line of the choice under the cursor shows.
func (m Model) follow() Model {
	body, spans := m.body(theme.Default())
	if m.cursor < len(spans) {
		m.top = pane.FollowLines(m.top, spans[m.cursor].from, spans[m.cursor].to, len(body), m.room())
	}
	return m
}

// Wheel scrolls the body by delta lines; the cursor stays where it is.
//
// Parameters: delta is the lines to move, negative for up.
// Returns: the dialog with the body scrolled.
func (m Model) Wheel(delta int) Model {
	body, _ := m.body(theme.Default())
	m.top = pane.ClampTop(m.top+delta, len(body), m.room())
	return m
}

// span is the lines a choice takes in the body, first included, last excluded.
type span struct{ from, to int }

// body is the wrapped text lines of the dialog and, for each choice, the lines
// it takes: Lines draws the first and Click maps a row to a choice with the
// second, so what is drawn and what is clickable never differ.
func (m Model) body(th theme.Theme) (body []string, spans []span) {
	// One cell less than the inside: the last cell is the scrollbar's when the
	// body scrolls, so no wrapped line loses its last letter under it.
	inner := m.rect.W - 5
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
		r := choicerow.Row{Label: c.label, Hint: c.hint, Selected: i == m.cursor, Hovered: i == m.over-1 && i != m.cursor}
		from := len(body)
		body = append(body, choicerow.Lines(th, r, inner)...)
		spans = append(spans, span{from, len(body)})
	}
	body = append(body, "")
	body = append(body, pane.Wrap(th.Muted(m.foot), inner, 0)...)
	return body, spans
}

// Click handles a left press: a press on a choice picks it, as Enter does on
// it, like a click in every single-choice list. A press elsewhere changes
// nothing.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen cells.
// Returns: the new dialog and what the host must do.
func (m Model) Click(th theme.Theme, x, y int) (Model, Outcome) {
	if i := m.choiceAt(th, x, y); i >= 0 {
		m.cursor = i
		return m.Key("enter")
	}
	return m, Outcome{}
}

// Hover notes the choice under the pointer at (x, y), drawn in the row hover look.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen cells.
// Returns: the dialog with the hover set.
func (m Model) Hover(th theme.Theme, x, y int) Model {
	m.over = m.choiceAt(th, x, y) + 1
	return m
}

// choiceAt is the choice drawn at the screen cell (x, y), or -1.
func (m Model) choiceAt(th theme.Theme, x, y int) int {
	if !m.rect.Contains(x, y) {
		return -1
	}
	body, spans := m.body(th)
	line := pane.BodyLine(y-m.rect.Y-1, m.top, len(body), m.room())
	sp := make([]choicerow.Span, len(spans))
	for i, s := range spans {
		sp[i] = choicerow.Span{From: s.from, To: s.to}
	}
	return choicerow.At(sp, line)
}
