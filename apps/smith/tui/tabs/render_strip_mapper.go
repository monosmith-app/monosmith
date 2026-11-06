package tabs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Render draws the strip for the props: one row of buttons, a blank cell between them.
func Render(p Props) Strip {
	if p.Width < 1 {
		return Strip{}
	}
	return renderFlat(p)
}

// gapW is the blank cell between two tabs: the buttons share one resting
// background, so without it they would run into one block. It replaces the pipe.
const gapW = 1

// activeIndex is the position of the active tab, 0 when none is.
func activeIndex(p Props) int {
	for i, t := range p.Tabs {
		if t.ID == p.Active {
			return i
		}
	}
	return 0
}

// state is how a tab's button looks: the active one pressed in, the one under
// the pointer lit, the others at rest.
func state(id string, p Props) button.State {
	switch id {
	case p.Active:
		return button.Active
	case p.Hover:
		return button.Hover
	}
	return button.Rest
}

// renderFlat is the one-row strip: a button " Name (n) " for each tab with a
// blank cell between them, the active one pressed in. A badge follows its button.
func renderFlat(p Props) Strip {
	th := p.Theme
	avail := max(p.Width-1, 0)
	short := shortened(th, p.Tabs, p.ShortenOrder, avail, func(inner int) int { return inner + gapW })
	widths := make([]int, len(p.Tabs))
	for i, t := range p.Tabs {
		widths[i] = t.planned(th, short[t.ID]) + gapW
	}
	start, end := window(widths, activeIndex(p), avail)

	var b strings.Builder
	var spans []Span
	x := 0
	put := func(s string, w int) { b.WriteString(s); x += w }
	put(" ", 1)
	if start > 0 {
		spans = append(spans, Span{ID: Prev, X0: x, X1: x + markerW})
		put(th.Accent(Prev)+" ", markerW)
	}
	for i := start; i < end; i++ {
		t := p.Tabs[i]
		name := t.name(short[t.ID])
		label := button.Render(th, name, t.Key, state(t.ID, p))
		w := button.Width(th, name, t.Key)
		if t.Badge != "" {
			label += " " + t.Badge
			w += 1 + pane.Width(t.Badge)
		}
		spans = append(spans, Span{ID: t.ID, X0: x, X1: x + w})
		put(label, w)
		if i < end-1 {
			put(strings.Repeat(" ", gapW), gapW)
		}
	}
	if end < len(p.Tabs) {
		spans = append(spans, Span{ID: Next, X0: x, X1: x + markerW})
		put(th.Accent(Next), 1)
	}
	return Strip{Lines: []string{pane.Fit(b.String(), p.Width)}, Spans: spans}
}
