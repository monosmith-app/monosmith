package tabs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Render draws the strip for the props.
func Render(p Props) Strip {
	if p.Width < 1 {
		return Strip{}
	}
	if p.Tall {
		return renderTall(p)
	}
	return renderFlat(p)
}

// activeIndex is the position of the active tab, 0 when none is.
func activeIndex(p Props) int {
	for i, t := range p.Tabs {
		if t.ID == p.Active {
			return i
		}
	}
	return 0
}

// styleLabel is a tab's text in the style of its state: the active tab in
// activeStyle, the one under the pointer highlighted, the others dim.
func styleLabel(th theme.Theme, id string, p Props, label string, activeStyle func(string) string) string {
	switch id {
	case p.Active:
		return activeStyle(pane.Strip(label))
	case p.Hover:
		return th.Hover(label)
	}
	return th.Muted(label)
}

// renderFlat is the one-row strip: " Name " for each tab, the active one
// highlighted.
func renderFlat(p Props) Strip {
	th := p.Theme
	avail := max(p.Width-1, 0)
	short := shortened(p.Tabs, p.ShortenOrder, avail, func(inner int) int { return inner + 1 })
	widths := make([]int, len(p.Tabs))
	for i, t := range p.Tabs {
		widths[i] = t.planned(short[t.ID]) + 1
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
		label := t.text(short[t.ID])
		w := pane.Width(label)
		label = styleLabel(th, t.ID, p, label, th.Selected)
		spans = append(spans, Span{ID: t.ID, X0: x, X1: x + w})
		put(label+" ", w+1)
	}
	if end < len(p.Tabs) {
		spans = append(spans, Span{ID: Next, X0: x, X1: x + markerW})
		put(th.Accent(Next), 1)
	}
	return Strip{Lines: []string{pane.Fit(b.String(), p.Width)}, Spans: spans}
}

// renderTall is the three-row strip:
//
//	╭──────────╮╭─────╮
//	│ Projects ││ Run │
//	│          └┴─────┴───────────────╮
//
// The active tab has no floor, so it opens into the pane below; the last row
// is that pane's top edge.
func renderTall(p Props) Strip {
	th := p.Theme
	// The last column holds the corner of the pane.
	avail := max(p.Width-1, 0)
	short := shortened(p.Tabs, p.ShortenOrder, avail, func(inner int) int { return inner + 2 })
	widths := make([]int, len(p.Tabs))
	for i, t := range p.Tabs {
		widths[i] = t.planned(short[t.ID]) + 2
	}
	start, end := window(widths, activeIndex(p), avail)

	top, mid := newCells(p.Width), newCells(p.Width)
	base := newCells(p.Width)
	for x := range base.ch {
		base.ch[x] = '─'
	}
	base.ch[0], base.ch[p.Width-1] = '├', '╮'
	if p.Width == 1 {
		base.ch[0] = '╮'
	}

	var spans []Span
	x := 0
	if start > 0 {
		mid.text(x, th.Accent(Prev)+" ", markerW)
		spans = append(spans, Span{ID: Prev, X0: x, X1: x + markerW})
		x += markerW
	}
	for i := start; i < end; i++ {
		t := p.Tabs[i]
		label := t.text(short[t.ID])
		w := pane.Width(label)
		if x+w+2 > p.Width-1 {
			break
		}
		active := t.ID == p.Active
		edge := func(s string) string { return th.Border(s, active) }
		top.text(x, edge("╭"+strings.Repeat("─", w)+"╮"), w+2)
		styled := styleLabel(th, t.ID, p, label, th.Accent)
		mid.text(x, edge("│")+styled+edge("│"), w+2)
		if active {
			base.set(x, '┘', true)
			if x == 0 {
				base.set(x, '│', true)
			}
			for dx := 1; dx <= w; dx++ {
				base.set(x+dx, ' ', false)
			}
			base.set(x+w+1, '└', true)
		} else {
			base.set(x+w+1, '┴', false)
		}
		spans = append(spans, Span{ID: t.ID, X0: x, X1: x + w + 2})
		x += w + 2
	}
	if end < len(p.Tabs) && x+markerW <= p.Width-1 {
		mid.text(x, " "+th.Accent(Next), markerW)
		spans = append(spans, Span{ID: Next, X0: x, X1: x + markerW})
		x += markerW
	}
	if t := p.Title; t != "" && pane.Width(t)+3 <= p.Width-x {
		at := p.Width - 1 - pane.Width(t) - 1
		mid.text(at, th.Muted(t), pane.Width(t))
	}
	if d := p.Divider; d > 0 && d < p.Width-1 && base.ch[d] == '─' {
		base.set(d, '┬', false)
	}
	return Strip{Lines: []string{top.render(th, p.Width, false), mid.render(th, p.Width, false), base.render(th, p.Width, true)}, Spans: spans}
}
