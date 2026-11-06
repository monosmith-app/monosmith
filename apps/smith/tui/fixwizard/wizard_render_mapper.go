package fixwizard

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// action is a labelled key on the bottom row; a click presses its key, which is also the one drawn in parentheses.
type action struct{ label, key string }

// buttons are the buttons of the phase, left to right.
func (w Wizard) buttons() []action {
	var out []action
	switch w.phase {
	case Asking:
		if w.canGoBack() || w.style.Back {
			out = append(out, action{words.Say("fixwizard.button.back"), "b"})
		}
		out = append(out, action{words.Say("fixwizard.button.cancel"), "esc"})
	case Reviewing:
		out = append(out, action{w.style.apply(), "enter"})
		if w.hasDiff {
			label := words.Say("fixwizard.button.show")
			if w.full {
				label = words.Say("fixwizard.button.hide")
			}
			out = append(out, action{label, "d"})
		}
		if w.style.Decline != "" {
			out = append(out, action{w.style.Decline, "n"})
		}
		if w.style.Skip != "" {
			out = append(out, action{w.style.Skip, "s"})
		}
		if w.canGoBack() || w.style.Back {
			out = append(out, action{words.Say("fixwizard.button.back"), "b"})
		}
		out = append(out, action{words.Say("fixwizard.button.cancel"), "esc"})
	case Applying:
		if w.logOpen {
			out = append(out, action{words.Say("fixwizard.button.closelog"), "l"})
		}
		out = append(out, action{words.Say("fixwizard.button.stop"), "ctrl+c"})
	case Failed:
		if !w.style.NoRetry {
			out = append(out, action{words.Say("fixwizard.button.retry"), "r"})
		}
		out = append(out, action{words.Say("fixwizard.button.copy"), "y"}, action{words.Say("fixwizard.button.close"), "esc"})
	}
	return out
}

// buttonSpans are the columns (from the dialog's left edge) each button covers.
func (w Wizard) buttonSpans(inner int) [][2]int { return w.spansIn(theme.Default(), inner) }

// spansIn is buttonSpans for the theme the dialog is drawn with: without
// colour every button is two cells wider, for its brackets.
func (w Wizard) spansIn(th theme.Theme, inner int) [][2]int {
	var labels, keys []string
	for _, b := range w.buttons() {
		labels, keys = append(labels, b.label), append(keys, b.key)
	}
	return button.RowSpans(th, labels, keys, inner)
}

// width is the dialog's width on the screen it was placed on.
func (w Wizard) width() int {
	if w.sw < 30 {
		return max(w.sw, 6)
	}
	return min(max(w.sw-6, 30), 100)
}

// visible is how many body rows the dialog shows.
func (w Wizard) visible() int { return max(w.Rect().H-4, 1) }

// Rect is where the dialog sits.
func (w Wizard) Rect() layout.Rect {
	width := w.width()
	h := max(w.sh-2, 5)
	if !w.LogOpen() {
		rows, _ := w.rows(theme.Default(), 0)
		h = min(len(rows)+4, h)
	}
	return layout.Rect{X: max((w.sw-width)/2, 0), Y: max((w.sh-h)/2, 0), W: width, H: h}
}

func (w Wizard) bodyRows() []string {
	rows, _ := w.rows(theme.Default(), 0)
	return rows
}

// title is set into the top edge.
func (w Wizard) title() string {
	switch w.phase {
	case Asking:
		n, total := len(w.answers)+1, len(w.answers)+len(w.resolved.Choices)
		if total == 1 && w.style.Ask != "" {
			return words.Say("fixwizard.title.part", "name", w.style.name(), "what", w.style.Ask)
		}
		return words.Say("fixwizard.title.asking", "name", w.style.name(), "what", w.style.ask(), "number", n, "total", total)
	case Reviewing:
		return words.Say("fixwizard.title.part", "name", w.style.name(), "what", w.style.review())
	case Applying:
		if w.logOpen {
			return words.Say("fixwizard.title.log", "verb", w.style.verb())
		}
		return w.style.verb()
	}
	return words.Say("fixwizard.title.failed", "name", w.style.name())
}

// rows is the whole body, before it is windowed, and for Asking the line each
// option is on.
func (w Wizard) rows(th theme.Theme, tick int) (rows []string, optionRow []int) {
	inner := max(w.width()-4, 4)
	switch w.phase {
	case Asking:
		return w.askingRows(th, inner)
	case Reviewing:
		return reviewLines(th, strings.Join(w.reviewShown(), "\n"), inner), nil
	case Applying:
		if w.logOpen {
			return w.logRows(th, inner), nil
		}
		return w.applyingRows(th, tick, inner), nil
	}
	rows = pane.Wrap(th.Failure(w.headline), inner, 2)
	if len(w.output) > 0 {
		rows = append(rows, "", th.Muted(words.Say("fixwizard.failed.printed")))
		for _, l := range w.output {
			rows = append(rows, pane.Wrap("  "+l, inner, 4)...)
		}
	}
	return rows, nil
}

func (w Wizard) askingRows(th theme.Theme, inner int) (rows []string, optionRow []int) {
	rows = append(rows, pane.Wrap(th.Muted(w.Fix.Description), inner, 0)...)
	if scope := doctorfix.Scope(w.Fix.AppliesTo); scope != "" {
		rows = append(rows, th.Muted(scope))
	}
	rows = append(rows, "")
	rows = append(rows, pane.Wrap(th.Bold(w.question.Question), inner, 0)...)
	rows = append(rows, "")
	for i, o := range w.question.Options {
		label := fmt.Sprint(i+1) + ") " + strings.ReplaceAll(o.Label, "{name}", w.text)
		if i == 0 {
			label += words.Say("fixwizard.suggested")
		}
		mark := choicerow.Cursor(th, i == w.cursor)
		optionRow = append(optionRow, len(rows))
		// A long label wraps under its number, so no option is cut off.
		for _, part := range pane.Wrap(mark+label, inner, 5) {
			if i == w.cursor {
				part = th.Selected(pane.Fit(part, inner))
			}
			rows = append(rows, part)
		}
	}
	// The row after the last option closes the last one's extent.
	optionRow = append(optionRow, len(rows))
	hint := words.Say("fixwizard.hint.pick")
	if w.style.Input != "" {
		field := w.text
		if w.editing {
			field += "▏"
			hint = words.Say("fixwizard.hint.typing", "field", strings.ToLower(w.style.Input))
		} else {
			hint = words.Say("fixwizard.hint.edit", "field", strings.ToLower(w.style.Input))
		}
		rows = append(rows, "", th.Bold(words.Say("fixwizard.field", "name", w.style.Input))+field)
	}
	rows = append(rows, "", th.Muted(hint))
	return rows, optionRow
}

func (w Wizard) applyingRows(th theme.Theme, tick, inner int) []string {
	headline := w.resolved.Description
	if w.busy != "" {
		headline = w.busy
	}
	rows := pane.Wrap(th.Bold(headline), inner, 0)
	rows = append(rows, "")
	spin := th.Accent(th.Spin(tick)) + " "
	waited := w.elapsed()
	switch {
	case w.total == 0 && w.step == 0:
		rows = append(rows, spin+progress.Waiting(w.style.working()+"…", waited))
	case w.step == 0:
		rows = append(rows, spin+progress.Waiting(words.Say("fixwizard.applying.then", "working", w.style.working(), "total", w.total), waited))
	case w.total == 0:
		// A job that does not know how many steps it has: the name alone.
		rows = append(rows, spin+progress.Waiting(capitalFirst(w.stepName)+"…", waited))
	default:
		rows = append(rows, spin+progress.Waiting(words.Say("fixwizard.applying.step", "step", w.step, "total", max(w.total, w.step), "name", w.stepName), waited))
	}
	// The detail rows are always there, so the dialog keeps its height.
	for i := 0; i < NoteRows; i++ {
		line := ""
		if i < len(w.note) {
			line = w.note[i]
		}
		rows = append(rows, th.Muted("  "+ansi.Truncate(line, max(inner-2, 1), "…")))
	}
	rows = append(rows, stepListRows(th, w.resolved.Steps, w.step, inner)...)
	return append(rows, "", th.Muted(applyingHint))
}

// logRows is the full log, a line to a row; a long line is cut, not wrapped, so
// a row of the view is a line of the log.
func (w Wizard) logRows(th theme.Theme, inner int) []string {
	if len(w.log) == 0 {
		return []string{th.Muted(words.Say("fixwizard.log.empty"))}
	}
	rows := make([]string, len(w.log))
	for i, l := range w.log {
		rows[i] = ansi.Truncate(l, inner, "…")
	}
	return rows
}

// onLogHint is true when the point is on the words "l full log" of the hint row.
func (w Wizard) onLogHint(x, y int) bool {
	r := w.Rect()
	rows := w.bodyRows()
	if len(rows) > w.visible() || y != r.Y+len(rows) {
		return false
	}
	from := r.X + 2 + ansi.StringWidth(applyingHint) - ansi.StringWidth(logHint)
	return x >= from && x < from+ansi.StringWidth(logHint)
}

// applyingHint tells what the keys do while a step runs, so a key that waits
// is not mistaken for a key that was lost.
var applyingHint = words.Say("fixwizard.hint.applying", "log", logHint)

// logHint is the end of the applying hint: the key that opens the full log.
var logHint = words.Say("fixwizard.hint.log")

// follow scrolls the question so the option under the cursor shows, clear of
// the "more" markers.
func (w Wizard) follow() Wizard {
	if w.phase != Asking {
		return w
	}
	_, at := w.rows(theme.Default(), 0)
	if w.cursor+1 >= len(at) {
		return w
	}
	row, last, vis := at[w.cursor], at[w.cursor+1]-1, w.visible()
	if row < w.top+1 {
		w.top = max(row-1, 0)
	}
	if last > w.top+vis-2 {
		w.top = last - vis + 2
	}
	return w
}

// optionAt is the option drawn on screen row y.
func (w Wizard) optionAt(y int) (int, bool) {
	if w.phase != Asking {
		return 0, false
	}
	r := w.Rect()
	rows, at := w.rows(theme.Default(), 0)
	line := y - r.Y - 1
	if line < 0 || line >= w.visible() {
		return 0, false
	}
	if len(rows) > w.visible() {
		line += pane.ClampTop(w.top, len(rows), w.visible())
	}
	for i := 0; i+1 < len(at); i++ {
		if line >= at[i] && line < at[i+1] {
			return i, true
		}
	}
	return 0, false
}

// TextAt is the plain text of the body line drawn on screen row y, or "" when
// no body line is there. A caller uses it to give a line of the review a click
// target, such as a tick box.
func (w Wizard) TextAt(y int) string {
	rows := w.bodyRows()
	line := y - w.Rect().Y - 1
	if line < 0 || line >= w.visible() {
		return ""
	}
	if len(rows) > w.visible() {
		line += pane.ClampTop(w.top, len(rows), w.visible())
	}
	if line >= len(rows) {
		return ""
	}
	return pane.Strip(rows[line])
}

// Lines draws the dialog: Rect().H lines of Rect().W cells. tick moves the
// spinner while a step runs.
func (w Wizard) Lines(th theme.Theme, tick int) []string {
	r := w.Rect()
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	rows, _ := w.rows(th, tick)
	total := len(rows)
	rows, _ = pane.WindowMarked(th, rows, nil, w.top, w.visible())
	rows = pane.Scrollbar(th, rows, inner, pane.ClampTop(w.top, total, w.visible()), total)
	out := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, inner) }
	for i, l := range rows {
		if i >= r.H-4 {
			break
		}
		out = append(out, row(l))
	}
	for len(out) < r.H-4 {
		out = append(out, row(""))
	}
	out = append(out, row(""))
	line := ""
	spans := w.spansIn(th, inner)
	col := 0
	for i, b := range w.buttons() {
		line += strings.Repeat(" ", spans[i][0]-2-col)
		state := button.Rest
		switch {
		case i == w.hover:
			state = button.Hover
		case i == 0 && (w.phase == Reviewing || w.phase == Failed):
			state = button.Active
		}
		line += button.Render(th, b.label, b.key, state)
		col = spans[i][1] - 2
	}
	out = append(out, row(line))
	return pane.Frame(th, w.title(), out, r.W, r.H, true)
}

// reviewShown is the review text as it is drawn: a line for each changed file,
// or the whole diff once the reader asks for it.
func (w Wizard) reviewShown() []string {
	if w.full || !w.hasDiff {
		return w.review
	}
	return w.summary
}
