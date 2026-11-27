package settingsview

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// selectedMark is the mark of the selected section, as every list of the TUI draws it.
const selectedMark = "▶ "

// The ids of the two buttons under the fields of a setting.
const (
	saveButton  = "save"
	resetButton = "reset"
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
	// btnRow is the row of the Save and Reset buttons of the right pane (-1 when
	// none is drawn), and btns where each one sits on it.
	btnRow int
	btns   []buttonSpan
}

// buttonSpan is a button on its row: the cells [from, to) and its id.
type buttonSpan struct {
	from, to int
	id       string
}

// paneLines is what one pane of the tab draws: its lines and what its rows stand for.
type paneLines struct {
	lines   []string
	rowItem map[int]int
	btnRow  int
	btns    []buttonSpan
}

// View draws the tab in w by h cells: the list of sections at the left and the
// open section at the right, for a host that has one pane for the tab. px, py
// are where the pointer is, in the same cells (-1 when it is elsewhere): the
// row under it is highlighted.
func (m Model) View(th theme.Theme, w, h, px, py int) []string {
	return m.frame(th, w, h, px, py).lines
}

func (m Model) frame(th theme.Theme, w, h, px, py int) frame {
	f := frame{rowItem: map[int]int{}, rowSection: map[int]Section{}, btnRow: -1}
	if w < 1 || h < 1 {
		f.lines = pane.Block(nil, max(w, 0), max(h, 0))
		return f
	}
	f.listW = m.sectionsWidth(w)
	rw := max(w-f.listW-1, 0)
	left, rowSection := m.sectionsLines(th, f.listW, h, px, py, false)
	f.rowSection = rowSection
	right := m.right(th, rw, h, px-f.listW-1, py)
	f.rowItem, f.btnRow = right.rowItem, right.btnRow
	for _, b := range right.btns {
		f.btns = append(f.btns, buttonSpan{b.from + f.listW + 1, b.to + f.listW + 1, b.id})
	}
	sep := th.Faint("│")
	out := make([]string, h)
	for i := range out {
		out[i] = pane.Fit(left[i], f.listW) + sep + pane.Fit(right.lines[i], rw)
	}
	f.lines = pane.Block(out, w, h)
	return f
}

// SectionsPane draws the list of sections alone in w by h cells, for the frame
// at the left of a wide screen: each section on its row, the selected one
// marked, and where the settings of a project are at the foot. px, py are the
// pointer in the same cells (-1 when it is elsewhere).
//
// Parameters: th is the theme; w and h are the size; px and py are the pointer.
// Returns: exactly h lines of w cells.
func (m Model) SectionsPane(th theme.Theme, w, h, px, py int) []string {
	lines, _ := m.sectionsLines(th, w, h, px, py, false)
	return pane.Block(lines, max(w, 0), max(h, 0))
}

// FieldsPane draws the open section alone in w by h cells, for the frame at
// the right of a wide screen: its rows (the settings of smith.toml as a table),
// the box of the selected row or the editor that is open, and the buttons of a
// setting. px, py are the pointer in the same cells (-1 when it is elsewhere).
//
// Parameters: th is the theme; w and h are the size; px and py are the pointer.
// Returns: exactly h lines of w cells.
func (m Model) FieldsPane(th theme.Theme, w, h, px, py int) []string {
	if w < 1 || h < 1 {
		return pane.Block(nil, max(w, 0), max(h, 0))
	}
	return pane.Block(m.right(th, w, h, px, py).lines, w, h)
}

// Narrow draws the tab for a narrow screen, one pane at a time: the list of
// sections, each with a line about it, while the keyboard is there; the open
// section (as FieldsPane draws it) once Enter opened it.
//
// Parameters: th is the theme; w and h are the size; px and py are the pointer.
// Returns: exactly h lines of w cells.
func (m Model) Narrow(th theme.Theme, w, h, px, py int) []string {
	if w < 1 || h < 1 {
		return pane.Block(nil, max(w, 0), max(h, 0))
	}
	if m.onItems {
		return m.FieldsPane(th, w, h, px, py)
	}
	lines, _ := m.sectionsLines(th, w, h, px, py, true)
	return pane.Block(lines, w, h)
}

// ToSections puts the keyboard on the list of sections, as Esc does; an open
// editor keeps it.
//
// Returns: the tab with the keyboard on its sections.
func (m Model) ToSections() Model {
	if m.ed == nil {
		m.onItems = false
	}
	return m
}

// OnSections is true while the keyboard is on the list of sections, false
// while it is in the rows of the open section.
func (m Model) OnSections() bool { return !m.onItems }

// PaneTitle is the title of the frame of the open section: its name, and the
// file it writes when that is smith.toml ("Workspace · smith.toml").
func (m Model) PaneTitle() string {
	if m.section.readsWorkspace() {
		return words.Say("settingspage.pane.title", "section", m.titleOf(m.section))
	}
	return m.titleOf(m.section)
}

// SectionsClick is a click at row y of the pane SectionsPane draws in w by h
// cells: it selects the section there, and a double click opens it.
//
// Parameters: w and h are the size of the pane; y is the row; double is true
// for a double click.
// Returns: the tab.
func (m Model) SectionsClick(w, h, y int, double bool) Model {
	return m.pickSection(w, h, y, double, false)
}

// FieldsClick is a click at (x, y) of the pane FieldsPane draws in w by h
// cells: it selects a row, picks a value of the open editor or presses a
// button; a double click on a row does what Enter does.
//
// Parameters: w and h are the size of the pane; x and y are the cell; double
// is true for a double click.
// Returns: the tab, and what the click asks the shell to do.
func (m Model) FieldsClick(w, h, x, y int, double bool) (Model, Result) {
	plain := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	p := m.right(plain, w, h, -1, -1)
	if y == p.btnRow {
		return m.pressButton(p.btns, x)
	}
	return m.clickItem(p.rowItem, y, double)
}

// NarrowClick is a click at (x, y) of what Narrow draws in w by h cells: a
// section of the list (a click on the selected one, or a double click, opens
// it), or what FieldsClick takes in an open section.
//
// Parameters: w and h are the size; x and y are the cell; double is true for a
// double click.
// Returns: the tab, and what the click asks the shell to do.
func (m Model) NarrowClick(w, h, x, y int, double bool) (Model, Result) {
	if m.onItems {
		return m.FieldsClick(w, h, x, y, double)
	}
	return m.pickSection(w, h, y, double, true), Result{}
}

// pickSection selects the section at row y of the list; open (or a second
// click on the selected one in the narrow list) opens it.
func (m Model) pickSection(w, h, y int, double, narrow bool) Model {
	if m.ed != nil {
		return m
	}
	plain := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	_, rowSection := m.sectionsLines(plain, w, h, -1, -1, narrow)
	if s, ok := rowSection[y]; ok {
		again := s == m.section
		m.section, m.onItems, m.note = s, double || (narrow && again), ""
	}
	return m
}

// pressButton is a click on the row of buttons at column x.
func (m Model) pressButton(btns []buttonSpan, x int) (Model, Result) {
	for _, b := range btns {
		if x < b.from || x >= b.to {
			continue
		}
		m.onItems = true
		if b.id == resetButton {
			return m.resetSetting()
		}
		return m.activate()
	}
	return m, Result{}
}

// sectionsLines draws the list of sections in w by h cells: a blank row, each
// section (with the line about it when subs is set), a blank row between them
// when they fit, and where the settings of a project are at the foot.
// rowSection maps each drawn row to its section.
func (m Model) sectionsLines(th theme.Theme, w, h, px, py int, subs bool) ([]string, map[int]Section) {
	rowSection := map[int]Section{}
	if w < 1 || h < 1 {
		return make([]string, max(h, 0)), rowSection
	}
	secs := m.sections()
	per := 1
	if subs {
		per = 2
	}
	hint := pane.Wrap(th.Muted(words.Say("settingspage.sections.hint")), max(w-1, 4), 0)
	need := 1 + len(secs)*per
	gap := need+len(secs)-1+len(hint)+1 <= h
	if gap {
		need += len(secs) - 1
	}
	lines := []string{}
	if h > len(secs)*per {
		lines = append(lines, "")
	}
	for i, s := range secs {
		if i > 0 && gap {
			lines = append(lines, "")
		}
		row := len(lines)
		title := m.titleOf(s)
		label := "  " + title
		switch {
		case s == m.section && !m.onItems:
			// The selected look of every list: the mark and the name in the accent on the selected background.
			label = th.Selected(pane.Fit(th.Accent(selectedMark)+th.Accent(th.Bold(title)), w))
		case s == m.section:
			label = th.Bold(selectedMark + title)
		case px >= 0 && px < w && (py == row || (subs && py == row+1)):
			label = th.RowHover(pane.Fit(label, w))
		}
		lines = append(lines, label)
		rowSection[row] = s
		if subs {
			rowSection[row+1] = s
			lines = append(lines, "  "+th.Muted(ansi.Truncate(m.subtitle(s), max(w-3, 1), "…")))
		}
	}
	if need+len(hint)+1 <= h {
		for len(lines) < h-len(hint)-1 {
			lines = append(lines, "")
		}
		lines = append(lines, hint...)
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return lines[:h], rowSection
}

// subtitle is the line about a section in the narrow list: what it holds now.
func (m Model) subtitle(s Section) string {
	switch {
	case s == Plugins && m.loaded:
		on := 0
		for _, p := range m.snap.Plugins {
			if p.Enabled {
				on++
			}
		}
		return words.Say("settingspage.sub.plugins", "on", on, "external", len(m.snap.External))
	case s == Workspace && m.loaded:
		set := 0
		for _, st := range m.snap.Settings {
			if st.Set {
				set++
			}
		}
		return words.Count("settingspage.sub.workspace", len(m.snap.Settings), "set", set)
	case s == Recipes && m.loaded:
		return words.Count("settingspage.sub.recipes", len(m.snap.Recipes))
	case s == Style:
		return words.Say("settingspage.sub.style", "theme", orAuto(m.savedTheme), "symbols", orAuto(m.savedIcons))
	case s == Architecture && m.loaded && m.snap.Architecture != "":
		return words.Say("settingspage.sub.architecture", "name", m.snap.Architecture)
	}
	return s.About()
}

// right is the right pane in w by h cells: the heading, then an editor or the
// rows and the box of the selected one.
func (m Model) right(th theme.Theme, w, h, px, py int) paneLines {
	p := m.rightRows(th, w, h, px, py)
	if m.busy != "" && h >= 3 {
		// A change is under way: the last row says so, with a spinner and the time after 2 s.
		p.lines[h-1] = " " + th.Accent(th.Spin(m.tick)) + " " + progress.Waiting(capital(m.busy)+"…", m.waited)
		if p.btnRow == h-1 {
			p.btnRow, p.btns = -1, nil
		}
	}
	return p
}

// capital puts the first letter of a sentence in capitals.
func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// tableSection is true for the sections whose rows are settings, drawn as a
// table with the Save and Reset buttons under the box.
func (m Model) tableSection() bool { return m.section == Workspace || m.section == Project }

func (m Model) rightRows(th theme.Theme, w, h, px, py int) paneLines {
	p := paneLines{lines: make([]string, h), rowItem: map[int]int{}, btnRow: -1}
	rows := p.lines
	rows[0] = " " + th.Muted(m.section.About())
	if m.section == Project {
		// The section of a project is drawn where no frame names it.
		rows[0] = " " + th.Title(m.titleOf(m.section)) + "  " + rows[0][1:]
	}
	body := h - 1
	if body < 1 {
		return p
	}
	put := func(at int, lines []string) {
		for i, l := range lines {
			if at+i < h {
				rows[at+i] = l
			}
		}
	}
	if m.ed != nil {
		m.editorPane(th, w, h, px, py, &p)
		return p
	}
	if lines, blocked := m.unavailable(th, w); blocked {
		put(1, lines)
		return p
	}
	list := m.items(m.section)
	cursor := m.cursor[m.section]
	head := 0
	if m.tableSection() && body >= 8 && len(list) > 0 {
		// The column heading and its rule stay at the top; the rows scroll under them.
		rows[1] = m.tableHead(th, w)
		rows[2] = th.Faint(" " + strings.Repeat("─", max(w-2, 0)))
		head = 2
	}
	all, lineOf := m.listLines(th, list, w, cursor)
	if len(all) == 0 {
		all = m.emptyList(th, w)
	}
	foot := 0
	if m.tableSection() && body-head >= 8 && cursor < len(list) && list[cursor].kind == kindSetting {
		foot = 1
	}
	listH, detailH, noteRow := m.rowSplit(body - head - foot)
	top := 0
	if cursor < len(lineOf) {
		top = m.scroll[m.section].Show(lineOf[cursor], listH, len(all))
	}
	at := 1 + head
	for i := 0; i < listH && top+i < len(all); i++ {
		rows[at+i] = all[top+i]
		for idx, line := range lineOf {
			if line != top+i {
				continue
			}
			p.rowItem[at+i] = idx
			if idx != cursor && px >= 0 && px < w && py == at+i {
				rows[at+i] = th.RowHover(pane.Fit(rows[at+i], w))
			}
		}
	}
	copy(rows[at:at+listH], pane.Scrollbar(th, rows[at:at+listH], w, top, len(all)))
	if detailH > 0 {
		put(at+listH, m.box(th, list, cursor, w, detailH+1))
	}
	if foot > 0 {
		p.btnRow = h - 1
		line, spans := m.buttons(th, list[cursor], px, py == h-1)
		rows[h-1], p.btns = line, spans
	}
	if noteRow > 0 {
		note := m.note
		switch {
		case note == "" && foot > 0:
			note = th.Muted(words.Say("settingspage.writes"))
		case note == "" && detailH == 0 && cursor < len(list):
			note = th.Muted(m.summary(list[cursor]))
		case m.failed:
			note = th.Failure(note)
		case note != "":
			note = th.Info(note)
		}
		// A note too long for the row ends with an ellipsis instead of being cut.
		rows[h-1-foot] = " " + ansi.Truncate(note, max(w-1, 1), "…")
	}
	return p
}

// box is the selected row told in a rounded box of h rows, w cells wide, with
// the row's name in its top border: what it is, what it does, and the command
// line it stands for, dim.
func (m Model) box(th theme.Theme, list []item, cursor, w, h int) []string {
	var lines []string
	title := ""
	if cursor < len(list) {
		it := list[cursor]
		title = m.boxTitle(it)
		lines = m.detail(th, it, w-4)
	}
	if h < 3 || w < 8 {
		return indent(lines)
	}
	inner := h - 2
	body := indent(lines)
	if hidden := len(body) - inner; hidden > 0 && inner >= 2 {
		// Say that the detail goes on instead of cutting it in the middle.
		body = append(body[:inner-1], " "+th.Muted(words.Say("settingsview.detail.more", "count", hidden+1)))
	}
	return pane.Frame(th, pane.CutEllipsis(title, max(w-6, 1)), body, w, h, m.onItems)
}

// boxTitle is the name in the top border of the box of a row.
func (m Model) boxTitle(it item) string {
	switch it.kind {
	case kindSetting:
		return m.settingAt(it).Path
	case kindInstalled:
		return m.snap.Installed[it.idx].Name
	case kindInstall:
		return words.Say("settingsview.install.title")
	case kindSearch:
		return words.Say("settingsview.search.title")
	case kindUpdate:
		return words.Say("settingsview.update.title")
	case kindUpdateMute:
		return words.Say("settingsview.update.mute")
	}
	return it.key
}

// buttons is the row of the Save and Reset buttons of a setting, from the left,
// and where each one sits. Save is the primary button; Reset is muted while
// the setting has no value written. hovered is true while the pointer is on
// the row.
func (m Model) buttons(th theme.Theme, it item, px int, hovered bool) (string, []buttonSpan) {
	st := m.settingAt(it)
	save, reset := words.Say("settingspage.button.save"), words.Say("settingspage.button.reset")
	saveW, resetW := button.Width(th, save, "enter"), button.Width(th, reset, "u")
	x := 1
	spans := []buttonSpan{{x, x + saveW, saveButton}, {x + saveW + 2, x + saveW + 2 + resetW, resetButton}}
	state := func(b buttonSpan, rest button.State) button.State {
		if hovered && px >= b.from && px < b.to && rest != button.Disabled {
			return button.Hover
		}
		return rest
	}
	resetState := button.Rest
	if !st.Set {
		resetState = button.Disabled
	}
	// Save is the primary button: Enter does what it does.
	saveState, resetState := state(spans[0], button.Active), state(spans[1], resetState)
	line := " " + button.Render(th, save, "enter", saveState) + "  " + button.Render(th, reset, "u", resetState)
	return line, spans
}

// tableHead is the column heading of the table of settings.
func (m Model) tableHead(th theme.Theme, w int) string {
	pw := m.pathWidth(w)
	cols := m.columns(w - 3 - pw)
	row := "   " + pane.Cell(words.Say("settingspage.column.setting"), pw)
	names := []string{words.Say("settingspage.column.value"), words.Say("settingspage.column.default"), words.Say("settingspage.column.type")}
	for i, cw := range cols {
		row += pane.Cell(names[i], cw)
	}
	return th.Muted(th.Bold(row))
}

// columns are the widths of the Value, Default and Type columns in the room
// left of the names: the type, then the default, leave first when it is short.
func (m Model) columns(room int) []int {
	switch {
	case room >= valueW+defaultW+typeW:
		return []int{max(room-defaultW-typeW, valueW), defaultW, typeW}
	case room >= valueW+defaultW:
		return []int{room - defaultW, defaultW}
	}
	return []int{max(room, 1)}
}

// The widths the columns of the table of settings want.
const (
	valueW   = 18
	defaultW = 14
	typeW    = 7
)

// editorPane draws the open editor in the right pane: the heading, the editor
// in a rounded box titled by what is changed, and its keys under the box.
func (m Model) editorPane(th theme.Theme, w, h, px, py int, p *paneLines) {
	body := h - 1
	if body < 5 || w < 12 {
		lines, rowOf := m.editorLines(th, w, body)
		for i, l := range lines {
			if 1+i < h {
				p.lines[1+i] = l
			}
		}
		m.markEditorRows(th, rowOf, 1, w, h, px, py, p)
		return
	}
	inner := body - 3
	lines, rowOf := m.editorLines(th, w-3, inner+1)
	keys := lines[len(lines)-1]
	box := pane.Frame(th, pane.CutEllipsis(m.ed.title, max(w-6, 1)), lines[:len(lines)-1], w, inner+2, true)
	for i, l := range box {
		p.lines[1+i] = l
	}
	p.lines[h-1] = keys
	m.markEditorRows(th, rowOf, 2, w, h, px, py, p)
}

// markEditorRows maps the editor's rows (rowOf, by line) to the rows of the
// pane, the first line at row at, and lights the one under the pointer.
func (m Model) markEditorRows(th theme.Theme, rowOf map[int]int, at, w, h, px, py int, p *paneLines) {
	for line, i := range rowOf {
		row := at + line
		if row >= h-1 {
			continue
		}
		p.rowItem[row] = i
		if i != m.ed.at && px >= 0 && px < w && py == row {
			p.lines[row] = th.RowHover(p.lines[row])
		}
	}
}

// sectionsWidth is the width of the list of sections in a tab w cells wide:
// the width the person gave it, else a fifth of the tab.
func (m Model) sectionsWidth(w int) int {
	switch {
	case w < 30:
		return max(w/3, 0)
	case m.listW > 0:
		return pane.Clamp(m.listW, minSections, max(w-minRight, minSections))
	}
	return pane.Clamp(w/5, minSections, 16)
}

// rowSplit divides the body of the right pane (the rows under its heading) into
// the rows of the list, the rows of the detail under it, and the row of the
// note. The detail has the rows the person gave it, else what its section
// wants; a body too short for both shows the list alone.
func (m Model) rowSplit(body int) (listH, detailH, noteRow int) {
	if body >= 5 {
		noteRow = 1
	}
	avail := body - noteRow
	if avail >= 9 {
		detailH = min(detailWant[m.section], avail-5)
		if m.detailRows > 0 {
			detailH = pane.Clamp(m.detailRows, minDetailRows, avail-5)
		}
	}
	listH = max(avail, 0)
	if detailH > 0 {
		listH = avail - detailH - 1
	}
	return listH, detailH, noteRow
}

// The least the list of sections, the right pane and the detail may be.
const (
	minSections   = 11
	minRight      = 24
	minDetailRows = 2
)

// detailWant is how many rows the detail of each section would like.
var detailWant = map[Section]int{Plugins: 8, Architecture: 10, Style: 9, Workspace: 10, Recipes: 6, Assistants: 4, CI: 4, Forms: 4, Update: 7, Project: 9}

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
			row = th.Selected(pane.Fit(th.Accent(pane.Strip(row)), w))
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
	if !m.section.readsWorkspace() {
		return nil, false
	}
	if m.section == Project {
		switch {
		case m.projectLoading:
			return []string{"", " " + th.Muted(words.Say("settingsview.project.reading", "name", m.proj.Name))}, true
		case m.proj.Err != nil:
			return append([]string{""}, indent(pane.Wrap(th.Failure(m.proj.Err.Error()), w-2, 0))...), true
		}
		return nil, false
	}
	switch {
	case m.b.Load == nil:
		return []string{"", " " + th.Muted(words.Say("settingsview.unavailable"))}, true
	case !m.loaded:
		return []string{"", " " + th.Muted(words.Say("settingsview.reading"))}, true
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

// Click selects what is at (x, y) of the tab View draws: a section, a row or a
// button; a double click does what Enter does on a row.
func (m Model) Click(w, h, x, y int, double bool) (Model, Result) {
	plain := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	f := m.frame(plain, w, h, -1, -1)
	if x < f.listW {
		if m.ed == nil {
			if s, ok := f.rowSection[y]; ok {
				m.section, m.onItems = s, double
				m.note = ""
			}
		}
		return m, Result{}
	}
	if y == f.btnRow {
		return m.pressButton(f.btns, x)
	}
	return m.clickItem(f.rowItem, y, double)
}

// clickItem is a click on row y of the right pane, whose rows stand for the
// items of rowItem: a value of the open editor, or a row of the section.
func (m Model) clickItem(rowItem map[int]int, y int, double bool) (Model, Result) {
	i, ok := rowItem[y]
	if m.ed != nil {
		// Only the rows of a path finder take a click while a value is edited.
		switch {
		case ok && m.ed.finder != nil:
			e := *m.ed
			sel := e.finder.Select(i)
			e.finder = &sel
			m.ed = &e
			if double {
				return m.editorKey(tea.KeyMsg{Type: tea.KeyEnter})
			}
		case ok && m.ed.kind == editChoose:
			// A click on a value of a pick list picks it, as in every single-choice list.
			e := *m.ed
			e.at = i
			m.ed = &e
			return m.editorKey(tea.KeyMsg{Type: tea.KeyEnter})
		case ok && m.ed.kind == editMulti:
			// A click on a tick box row puts the cursor there and ticks or unticks it, as Space does.
			e := *m.ed
			e.at = i
			m.ed = &e
			return m.editorKey(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		}
		return m, Result{}
	}
	if ok {
		m.cursor[m.section], m.onItems = i, true
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
