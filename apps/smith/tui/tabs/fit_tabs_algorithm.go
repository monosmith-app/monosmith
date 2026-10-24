package tabs

import "github.com/monosmith-app/monosmith/apps/smith/tui/pane"

// markerW is the room a scroll marker takes: the arrow and a space.
const markerW = 2

// name is the tab's name, shortened or not.
func (t Tab) name(short bool) string {
	if short && t.Short != "" {
		return t.Short
	}
	return t.Label
}

// text is what goes inside the tab: padded name and badge.
func (t Tab) text(short bool) string {
	s := " " + t.name(short)
	if t.Badge != "" {
		s += " " + t.Badge
	}
	return s + " "
}

// planned is the width the fit calculation gives the inside of a tab: the
// room for its badge counts whether or not the badge is there.
func (t Tab) planned(short bool) int {
	w := pane.Width(" " + t.name(short) + " ")
	extra := t.Reserve
	if t.Badge != "" {
		extra = max(extra, pane.Width(t.Badge)+1)
	}
	return w + extra
}

// shortened says which tabs use their short name: as few as it takes for every
// tab to fit in avail cells, in the order of order. cost turns the inside
// width of a tab into the cells it takes in the strip. When even the short
// names do not fit, all of them are short and the strip scrolls.
func shortened(tabs []Tab, order []string, avail int, cost func(inner int) int) map[string]bool {
	short := map[string]bool{}
	total := func() int {
		n := 0
		for _, t := range tabs {
			n += cost(t.planned(short[t.ID]))
		}
		return n
	}
	if total() <= avail {
		return short
	}
	for _, id := range order {
		for _, t := range tabs {
			if t.ID == id && t.Short != "" {
				short[id] = true
			}
		}
		if total() <= avail {
			return short
		}
	}
	return short
}

// window is the run of tabs [start, end) to draw: it holds the active tab, is
// as wide as the room allows, and leaves two cells for a marker on each side
// that has tabs hidden behind it. widths are the cells each tab takes.
func window(widths []int, active, avail int) (start, end int) {
	n := len(widths)
	if n == 0 {
		return 0, 0
	}
	active = min(max(active, 0), n-1)
	start, end = active, active+1
	fits := func(a, b int) bool {
		used := 0
		for _, w := range widths[a:b] {
			used += w
		}
		if a > 0 {
			used += markerW
		}
		if b < n {
			used += markerW
		}
		return used <= avail
	}
	for {
		grew := false
		if start > 0 && fits(start-1, end) {
			start--
			grew = true
		}
		if end < n && fits(start, end+1) {
			end++
			grew = true
		}
		if !grew {
			return start, end
		}
	}
}
