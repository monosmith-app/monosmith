package tabs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

// Render draws the strip for the props: one row of tabs, each its key and its
// name ("1 Home  2 Projects"), two blank cells between them.
//
// Parameters: p is what the strip is drawn from.
// Returns: the strip, exactly p.Width cells wide, and where each tab sits.
func Render(p Props) Strip {
	if p.Width < 1 {
		return Strip{}
	}
	return renderFlat(p)
}

// gapW is the blank cells between two tabs.
const gapW = 2

// activeIndex is the position of the active tab, 0 when none is.
func activeIndex(p Props) int {
	for i, t := range p.Tabs {
		if t.ID == p.Active {
			return i
		}
	}
	return 0
}

// look draws a tab's text in its state: the open view in the accent, bold
// and underlined; the one under the pointer on the hover background; the
// others with the key muted and the name plain.
func look(p Props, t Tab, name string) string {
	switch t.ID {
	case p.Active:
		return p.Theme.ActiveTab(t.text(name))
	case p.Hover:
		return p.Theme.RowHover(t.text(name))
	}
	if t.Key == "" {
		return name
	}
	return p.Theme.Muted(t.Key) + " " + name
}

// renderFlat is the one-row strip: "key name" for each tab with two blank
// cells between them, the open one in the active look. A badge follows its name.
func renderFlat(p Props) Strip {
	th := p.Theme
	avail := p.Width
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
	if start > 0 {
		spans = append(spans, Span{ID: Prev, X0: x, X1: x + markerW})
		put(th.Accent(Prev)+" ", markerW)
	}
	for i := start; i < end; i++ {
		t := p.Tabs[i]
		name := t.name(short[t.ID])
		label := look(p, t, name)
		w := pane.Width(t.text(name))
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
		put(" ", 1)
		spans = append(spans, Span{ID: Next, X0: x, X1: x + 1})
		put(th.Accent(Next), 1)
	}
	return Strip{Lines: []string{pane.Fit(b.String(), p.Width)}, Spans: spans}
}
