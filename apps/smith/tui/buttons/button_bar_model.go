package buttons

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Button is one control.
type Button struct {
	ID    string
	Label string
	// Icon is the symbol before the label; Busy replaces it with the spinner.
	Icon     string
	Shortcut string
	Disabled bool
	Busy     bool
}

// Bar is a row of buttons.
type Bar []Button

// gap is the space between buttons.
const gap = 2

func (b Button) text(th theme.Theme, tick int) string {
	icon := b.Icon
	if b.Busy {
		icon = th.Spin(tick)
	}
	s := b.Label
	if icon != "" {
		s = icon + " " + s
	}
	if b.Shortcut != "" {
		s += " " + b.Shortcut
	}
	return "[" + s + "]"
}

// span is the columns each visible button takes within w cells.
func (bar Bar) spans(th theme.Theme, w int) (starts, widths []int) {
	x := 0
	for _, b := range bar {
		n := pane.Width(b.text(th, 0))
		if x+n > w {
			break
		}
		starts, widths = append(starts, x), append(widths, n)
		x += n + gap
	}
	return starts, widths
}

// At is the ID of the button at column x of the row, if the click is on one
// and it is enabled.
func (bar Bar) At(th theme.Theme, w, x int) (id string, ok bool) {
	starts, widths := bar.spans(th, w)
	for i := range starts {
		if x >= starts[i] && x < starts[i]+widths[i] {
			return bar[i].ID, !bar[i].Disabled
		}
	}
	return "", false
}

// Render draws the row in w cells; hover is the ID of the button under the pointer.
func (bar Bar) Render(th theme.Theme, tick int, hover string, w int) string {
	starts, _ := bar.spans(th, w)
	var parts []string
	for i := range starts {
		b := bar[i]
		text := b.text(th, tick)
		switch {
		case b.Disabled:
			text = th.Muted(text)
		case b.Busy:
			text = th.Info(text)
		case b.ID == hover:
			text = th.Selected(text)
		default:
			text = th.Accent(text)
		}
		parts = append(parts, text)
	}
	return pane.Fit(strings.Join(parts, strings.Repeat(" ", gap)), w)
}
