package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// frame is the tab drawn in w by h cells, and what each row stands for.
type frame struct {
	lines []string
	// rowItem maps a row of the right pane to the index of its item; rowSection
	// maps a row of the list of sections to its section.
	rowItem    map[int]int
	rowSection map[int]Section
	// listW is the width of the list of sections.
	listW int
}

// View draws the tab in w by h cells. px, py are where the pointer is, in the
// same cells (-1 when it is elsewhere): the row under it is highlighted.
func (m Model) View(th theme.Theme, w, h, px, py int) []string {
	return m.frame(th, w, h, px, py).lines
}

func clamp(v, lo, hi int) int { return min(max(v, lo), max(lo, hi)) }

func (m Model) frame(th theme.Theme, w, h, px, py int) frame {
	f := frame{rowItem: map[int]int{}, rowSection: map[int]Section{}}
	if w < 1 || h < 1 {
		f.lines = pane.Block(nil, max(w, 0), max(h, 0))
		return f
	}
	f.listW = clamp(w/5, 11, 16)
	if w < 30 {
		f.listW = max(w/3, 0)
	}
	rw := max(w-f.listW-1, 0)
	left := make([]string, h)
	left[0] = th.Muted(" Sections")
	for i, s := range Sections {
		row := i + 1
		if row >= h {
			break
		}
		f.rowSection[row] = s
		label := "   " + s.Title()
		switch {
		case s == m.section && !m.onItems:
			label = th.Selected(pane.Fit(" "+th.Icons.Pointer+" "+s.Title(), f.listW))
		case s == m.section:
			label = th.Bold(" " + th.Icons.Pointer + " " + s.Title())
		case px >= 0 && px < f.listW && py == row:
			label = th.Hover(pane.Fit(label, f.listW))
		}
		left[row] = label
	}
	right, rowItem := m.right(th, rw, h, px-f.listW-1, py)
	f.rowItem = rowItem
	sep := th.Faint("│")
	out := make([]string, h)
	for i := range out {
		out[i] = pane.Fit(left[i], f.listW) + sep + pane.Fit(right[i], rw)
	}
	f.lines = pane.Block(out, w, h)
	return f
}

// right is the right pane in w by h cells: the heading, then an editor or the
// rows and the detail of the selected one.
func (m Model) right(th theme.Theme, w, h, px, py int) ([]string, map[int]int) {
	rows := make([]string, h)
	rowItem := map[int]int{}
	rows[0] = " " + th.Title(m.section.Title()) + "  " + th.Muted(m.section.About())
	body := h - 1
	if body < 1 {
		return rows, rowItem
	}
	put := func(at int, lines []string) {
		for i, l := range lines {
			if at+i < h {
				rows[at+i] = l
			}
		}
	}
	if m.ed != nil {
		put(1, m.editorLines(th, w, body))
		return rows, rowItem
	}
	if lines, blocked := m.unavailable(th, w); blocked {
		put(1, lines)
		return rows, rowItem
	}
	list := m.items(m.section)
	cursor := m.cursor[m.section]
	all, lineOf := m.listLines(th, list, w, cursor)
	if len(all) == 0 {
		all = m.emptyList(th, w)
	}
	noteRow := 0
	if body >= 5 {
		noteRow = 1
	}
	avail := body - noteRow
	detailH := 0
	if avail >= 9 {
		detailH = min(detailWant[m.section], avail-5)
	}
	listH := avail
	if detailH > 0 {
		listH = avail - detailH - 1
	}
	top := 0
	if cursor < len(lineOf) {
		top = m.scroll[m.section].Show(lineOf[cursor], listH, len(all))
	}
	for i := 0; i < listH && top+i < len(all); i++ {
		rows[1+i] = all[top+i]
		for idx, line := range lineOf {
			if line != top+i {
				continue
			}
			rowItem[1+i] = idx
			if idx != cursor && px >= 0 && px < w && py == 1+i {
				rows[1+i] = th.Hover(pane.Fit(pane.Strip(rows[1+i]), w))
			}
		}
	}
	copy(rows[1:1+listH], pane.Scrollbar(th, rows[1:1+listH], w, top, len(all)))
	if detailH > 0 {
		rows[1+listH] = th.Faint(" " + strings.Repeat("─", max(w-2, 0)))
		var detail []string
		if cursor < len(list) {
			detail = m.detail(th, list[cursor], w-2)
		} else {
			detail = m.emptyDetail(th, w-2)
		}
		for i := 0; i < detailH && i < len(detail); i++ {
			rows[2+listH+i] = " " + detail[i]
		}
	}
	if noteRow > 0 {
		note := m.note
		if note == "" && detailH == 0 && cursor < len(list) {
			note = th.Muted(m.summary(list[cursor]))
		} else if m.failed {
			note = th.Failure(note)
		} else if note != "" {
			note = th.Info(note)
		}
		rows[h-1] = " " + note
	}
	return rows, rowItem
}

// detailWant is how many rows the detail of each section would like.
var detailWant = map[Section]int{Plugins: 7, Style: 8, Workspace: 8, Recipes: 5, Forms: 3}

// listLines draws every row of the list with its group headings; lineOf is the
// line of each item. The row at the cursor is highlighted while the keyboard
// is in the rows.
func (m Model) listLines(th theme.Theme, list []item, w, cursor int) (lines []string, lineOf []int) {
	group := ""
	for i, it := range list {
		if it.group != group {
			group = it.group
			lines = append(lines, th.Muted(" "+group))
		}
		lineOf = append(lineOf, len(lines))
		row := m.itemRow(th, it, w)
		if i == cursor && strings.HasPrefix(row, "   ") {
			// The pointer marks the row even where there is no colour.
			row = " " + th.Icons.Pointer + row[2:]
		}
		switch {
		case i == cursor && m.onItems:
			row = th.Selected(pane.Fit(pane.Strip(row), w))
		case i == cursor:
			row = th.Bold(row)
		}
		lines = append(lines, pane.Fit(row, w))
	}
	return lines, lineOf
}

// unavailable explains why a section has no rows yet: not read, not readable,
// or no backend. blocked is true when it does.
func (m Model) unavailable(th theme.Theme, w int) (lines []string, blocked bool) {
	if m.section == Style || m.section == Forms {
		return nil, false
	}
	switch {
	case m.b.Load == nil:
		return []string{"", " " + th.Muted("The settings of this workspace are not available here.")}, true
	case !m.loaded:
		return []string{"", " " + th.Muted("Reading the workspace…")}, true
	case m.snap.Err != nil:
		lines = append([]string{""}, indent(pane.Wrap(th.Failure(m.snap.Err.Error()), w-2, 0))...)
		return lines, true
	}
	return nil, false
}

func indent(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = " " + l
	}
	return out
}

// Click selects what is at (x, y) of the body: a section, or a row; a double
// click does what Enter does on a row.
func (m Model) Click(w, h, x, y int, double bool) (Model, Result) {
	if m.ed != nil {
		return m, Result{}
	}
	plain := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	f := m.frame(plain, w, h, -1, -1)
	if x < f.listW {
		if s, ok := f.rowSection[y]; ok {
			m.section, m.onItems = s, double
			m.note = ""
		}
		return m, Result{}
	}
	if idx, ok := f.rowItem[y]; ok {
		m.cursor[m.section], m.onItems = idx, true
		m.note = ""
		if double {
			return m.activate()
		}
	}
	return m, Result{}
}

// Wheel scrolls the rows (or the review) by notches.
func (m Model) Wheel(dir int) Model {
	if m.ed == nil {
		// The list scrolls and the selection stays where it is.
		m.scroll[m.section].By(dir * 3)
		return m
	}
	key := map[bool]string{true: "up", false: "down"}[dir < 0]
	for range 3 {
		m, _ = m.editorKey(keyOf(key))
	}
	return m
}
