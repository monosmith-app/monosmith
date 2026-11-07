package doctorview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// FilterSpan is where a filter button sits on a headline line: the cells
// [From, To) answer a click.
type FilterSpan struct {
	Filter   Filter
	From, To int
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

// headlineRows draws the headline, which is the three filter buttons side by
// side from the second cell, and moves a button to the next line when it does
// not fit in w cells. spans[i] are the buttons of line i.
func (v Model) headlineRows(th theme.Theme, w int) (lines []string, spans [][]FilterSpan) {
	line, x := " ", 1
	var row []FilterSpan
	flush := func() {
		lines, spans = append(lines, line), append(spans, row)
		line, x, row = " ", 1, nil
	}
	for _, f := range filters {
		label, key := v.filterLabel(th, f), f.Key()
		bw := button.Width(th, label, key)
		if len(row) > 0 && x+bw > w {
			flush()
		}
		if len(row) > 0 {
			line += " "
			x++
		}
		row = append(row, FilterSpan{Filter: f, From: x, To: x + bw})
		line += button.Render(th, label, key, v.filterLook(f))
		x += bw
	}
	flush()
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, spans
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
