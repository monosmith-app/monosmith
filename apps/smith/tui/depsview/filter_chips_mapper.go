package depsview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// FilterSpan is where a chip sits on a line of the row: the cells [From, To)
// answer a click.
type FilterSpan struct {
	Filter   Filter
	From, To int
}

// Shows is true when the chip is the one that is on. A model with no chip
// chosen has the All chip on.
func (v Model) Shows(f Filter) bool {
	on := v.Filter
	if on == FilterNone {
		on = FilterAll
	}
	return on == f
}

// count is how many releases the chip's level holds among all that were found.
func (v Model) count(f Filter) int { return len(rowsAt(v.All, f)) }

// filterLabel is the text of a chip: the level and how many releases it holds.
func (v Model) filterLabel(f Filter) string {
	key := map[Filter]string{FilterAll: "depsview.filter.all", FilterPatch: "depsview.filter.patch", FilterMinor: "depsview.filter.minor", FilterMajor: "depsview.filter.major"}[f]
	return words.Say(key, "count", v.count(f))
}

// filterLook is how a chip looks: held down, under the pointer, on or off.
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

// headlineRows draws the chips side by side from the second cell, and moves a
// chip to the next line when it does not fit in w cells. spans[i] are the chips
// of line i.
func (v Model) headlineRows(th theme.Theme, w int) (lines []string, spans [][]FilterSpan) {
	line, x := " ", 1
	var row []FilterSpan
	flush := func() {
		lines, spans = append(lines, line), append(spans, row)
		line, x, row = " ", 1, nil
	}
	for _, f := range filters {
		label, key := v.filterLabel(f), f.Key()
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

// HeadlineSpans are the chips on each line at the top of the view, for the
// shell to turn a click into the key of the chip. There are none while the view
// has no release to filter.
func (v Model) HeadlineSpans(th theme.Theme, w int) [][]FilterSpan {
	if !v.Loaded || v.Err != nil || len(v.All) == 0 {
		return nil
	}
	_, spans := v.headlineRows(th, w)
	return spans
}
