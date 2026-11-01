package fixwizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// button is a labelled key on the bottom row; a click presses its key.
type button struct{ label, key string }

func spell(name, key string) string { return fmt.Sprintf("[ %s  %s ]", name, key) }

// buttons are the buttons of the phase, left to right.
func (w Wizard) buttons() []button {
	var out []button
	switch w.phase {
	case Asking:
		if w.canGoBack() || w.style.Back {
			out = append(out, button{spell("Back", "b"), "b"})
		}
		out = append(out, button{spell("Cancel", "Esc"), "esc"})
	case Reviewing:
		out = append(out, button{spell(w.style.apply(), "Enter"), "enter"})
		if w.hasDiff {
			label := "Show changes"
			if w.full {
				label = "Hide changes"
			}
			out = append(out, button{spell(label, "d"), "d"})
		}
		if w.style.Decline != "" {
			out = append(out, button{spell(w.style.Decline, "n"), "n"})
		}
		if w.style.Skip != "" {
			out = append(out, button{spell(w.style.Skip, "s"), "s"})
		}
		if w.canGoBack() || w.style.Back {
			out = append(out, button{spell("Back", "b"), "b"})
		}
		out = append(out, button{spell("Cancel", "Esc"), "esc"})
	case Applying:
		out = append(out, button{spell("Stop", "Ctrl+C"), "ctrl+c"})
	case Failed:
		if !w.style.NoRetry {
			out = append(out, button{spell("Retry", "r"), "r"})
		}
		out = append(out, button{spell("Copy output", "y"), "y"}, button{spell("Close", "Esc"), "esc"})
	}
	return out
}

// buttonSpans are the columns (from the dialog's left edge) each button covers.
func (w Wizard) buttonSpans(inner int) [][2]int {
	bs := w.buttons()
	total := 2 * max(len(bs)-1, 0)
	for _, b := range bs {
		total += ansi.StringWidth(b.label)
	}
	x := 2 + max((inner-total)/2, 0)
	var spans [][2]int
	for _, b := range bs {
		spans = append(spans, [2]int{x, x + ansi.StringWidth(b.label)})
		x += ansi.StringWidth(b.label) + 2
	}
	return spans
}

func (w Wizard) buttonAt(x, y int) int {
	r := w.Rect()
	if y != r.Y+r.H-2 {
		return -1
	}
	for i, s := range w.buttonSpans(r.W - 4) {
		if x-r.X >= s[0] && x-r.X < s[1] {
			return i
		}
	}
	return -1
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
func (w Wizard) Rect() dialog.Rect {
	width := w.width()
	rows, _ := w.rows(theme.Default(), 0)
	h := min(len(rows)+4, max(w.sh-2, 5))
	return dialog.Rect{X: max((w.sw-width)/2, 0), Y: max((w.sh-h)/2, 0), W: width, H: h}
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
			return w.style.name() + " · " + w.style.Ask
		}
		return fmt.Sprintf("%s · %s %d of %d", w.style.name(), w.style.ask(), n, total)
	case Reviewing:
		return w.style.name() + " · " + w.style.review()
	case Applying:
		return w.style.verb()
	}
	return w.style.name() + " failed"
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
		return w.applyingRows(th, tick, inner), nil
	}
	rows = pane.Wrap(th.Failure(w.headline), inner, 2)
	if len(w.output) > 0 {
		rows = append(rows, "", th.Muted("What it printed:"))
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
		label := fmt.Sprintf("%d) %s", i+1, strings.ReplaceAll(o.Label, "{name}", w.text))
		if i == 0 {
			label += "  (suggested)"
		}
		mark := "  "
		if i == w.cursor {
			mark = "▸ "
		}
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
	hint := "Up Down or a number picks · Enter chooses · Esc cancels"
	if w.style.Input != "" {
		field := w.text
		if w.editing {
			field += "▏"
			hint = "Type the " + strings.ToLower(w.style.Input) + " · Backspace deletes · Enter or Esc stops editing"
		} else {
			hint = "Up Down or a number picks · e edits the " + strings.ToLower(w.style.Input) + " · Enter chooses · Esc cancels"
		}
		rows = append(rows, "", th.Bold(w.style.Input+": ")+field)
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
	switch {
	case w.total == 0 && w.step == 0:
		rows = append(rows, th.Accent(th.Spin(tick))+" "+w.style.working()+"…")
	case w.step == 0:
		rows = append(rows, th.Accent(th.Spin(tick))+fmt.Sprintf(" %s, then %d step(s)…", w.style.working(), w.total))
	case w.total == 0:
		// A job that does not know how many steps it has: the name alone.
		rows = append(rows, th.Accent(th.Spin(tick))+" "+capitalFirst(w.stepName)+"…")
	default:
		rows = append(rows, th.Accent(th.Spin(tick))+fmt.Sprintf(" step %d of %d: %s", w.step, max(w.total, w.step), w.stepName))
	}
	rows = append(rows, stepListRows(th, w.resolved.Steps, w.step, inner)...)
	if e := elapsedText(w.elapsed()); e != "" {
		rows = append(rows, "", th.Muted(e))
	}
	return append(rows, "", th.Muted(applyingHint))
}

// applyingHint tells what the keys do while a step runs, so a key that waits
// is not mistaken for a key that was lost.
const applyingHint = "Ctrl+C or Esc stops the step that is running. Other keys do nothing until it is done."

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
		line += detail.ClampTop(w.top, len(rows), w.visible())
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
		line += detail.ClampTop(w.top, len(rows), w.visible())
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
	rows, _ = detail.Window(th, rows, nil, w.top, w.visible())
	rows = pane.Scrollbar(th, rows, inner, detail.ClampTop(w.top, total, w.visible()), total)
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
	spans := w.buttonSpans(inner)
	col := 0
	for i, b := range w.buttons() {
		line += strings.Repeat(" ", spans[i][0]-2-col)
		text := th.Muted(b.label)
		switch {
		case i == w.hover:
			text = th.Hover(b.label)
		case i == 0 && (w.phase == Reviewing || w.phase == Failed):
			text = th.Selected(b.label)
		}
		line += text
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
