package doctorview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// FilterSpan is where a button sits on a headline line: the cells [From, To)
// answer a click with the command ID. Filter is the filter of a filter button
// and FilterNone for the button that checks again.
type FilterSpan struct {
	Filter   Filter
	ID       string
	From, To int
}

// ReloadID is the command the button that checks again stands for.
const ReloadID = "view.reload"

// reloadLook is how the button that checks again looks: held down, under the
// pointer, or at rest.
func (v Model) reloadLook() button.State {
	switch {
	case v.PressedReload:
		return button.Active
	case v.UnderReload:
		return button.Hover
	}
	return button.Rest
}

// filterLabel is the text of a filter button: the mark of its state, how many
// checks are in it and the state.
func (v Model) filterLabel(th theme.Theme, f Filter) string {
	r := v.Report
	switch f {
	case FilterFailed:
		return words.Say("doctor.filter.failed", "icon", th.Icons.Failed, "count", r.Failed)
	case FilterWarnings:
		return words.Count("doctor.filter.warnings", r.Warnings, "icon", th.Icons.Warning)
	}
	return words.Say("doctor.filter.passed", "icon", th.Icons.Passed, "count", r.Passed)
}

// filterLook is how a filter button looks: held down, under the pointer, on
// (the selected look) or off (the rest look).
func (v Model) filterLook(f Filter) button.State {
	switch {
	case v.Pressed == f:
		return button.Active
	case v.Under == f:
		return button.Hover
	case v.Shows(f):
		return button.Active
	}
	return button.Rest
}

// headlineRows draws the headline, which is the three filter buttons and the
// button that checks again side by side from the second cell, and moves one to
// the next line when it does not fit in w cells. While doctor checks again, a
// hint takes the place of the last button. spans[i] are the buttons of line i.
func (v Model) headlineRows(th theme.Theme, w int) (lines []string, spans [][]FilterSpan) {
	line, x := " ", 1
	var row []FilterSpan
	flush := func() {
		lines, spans = append(lines, line), append(spans, row)
		line, x, row = " ", 1, nil
	}
	add := func(text string, bw int, span *FilterSpan) {
		if len(row) > 0 && x+bw > w {
			flush()
		}
		if len(row) > 0 {
			line += " "
			x++
		}
		if span != nil {
			span.From, span.To = x, x+bw
			row = append(row, *span)
		} else {
			// Not a button: a placeholder keeps the next one apart.
			row = append(row, FilterSpan{From: x, To: x})
		}
		line += text
		x += bw
	}
	for _, f := range filters {
		label, key := v.filterLabel(th, f), f.Key()
		add(button.Render(th, label, key, v.filterLook(f)), button.Width(th, label, key), &FilterSpan{Filter: f, ID: f.CommandID()})
	}
	if v.Loading {
		hint := words.Say("doctor.rechecking")
		add(th.Muted(hint), pane.Width(hint), nil)
	} else {
		label, key := words.Say("doctor.reload"), commands.ShortcutOf(ReloadID)
		add(button.Render(th, label, key, v.reloadLook()), button.Width(th, label, key), &FilterSpan{ID: ReloadID})
	}
	flush()
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, spans
}

// HeadHeight is how many lines the headline takes at the top of the page, with
// the blank line under it: the lines that stay while the checks scroll. It is 0
// while there is no list.
func (v Model) HeadHeight(th theme.Theme, w int) int {
	if v.Report == nil || v.Err != nil {
		return 0
	}
	head, _ := v.headlineRows(th, w)
	return len(head) + 1
}

// HeadlineSpans are the filter buttons on each line of the headline, for the
// shell to turn a click into the key of the filter.
func (v Model) HeadlineSpans(th theme.Theme, w int) [][]FilterSpan {
	if v.Report == nil || v.Err != nil {
		return nil
	}
	_, spans := v.headlineRows(th, w)
	return spans
}
