package settingsview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// ProjectPane draws the Project section alone in w by h cells, for the
// Settings tab of the detail of a project: its heading, its rows and the
// detail of the selected row, or the editor that is open. px, py are where
// the pointer is in the same cells (-1 when it is elsewhere). It is empty
// until OpenProject was called.
//
// Parameters: th is the theme; w and h are the size; px and py are the pointer.
// Returns: exactly h lines.
func (m Model) ProjectPane(th theme.Theme, w, h, px, py int) []string {
	if m.project == nil || w < 1 || h < 1 {
		return pane.Block(nil, max(w, 0), max(h, 0))
	}
	m.section, m.onItems, m.sideDetail = Project, true, false
	return pane.Block(m.right(th, w, h, px, py).lines, w, h)
}

// ProjectClick is a click at row y of the pane ProjectPane draws in w by h
// cells: it selects a row, or picks a value of the open editor; a double
// click does what Enter does. It knows no column, so it presses no button:
// ProjectClickAt does.
//
// Parameters: w and h are the size of the pane; y is the row; double is true
// for a double click.
// Returns: the tab, and what the click asks the shell to do.
func (m Model) ProjectClick(w, h, y int, double bool) (Model, Result) {
	return m.ProjectClickAt(w, h, -1, y, double)
}

// ProjectClickAt is ProjectClick at the cell (x, y): a click on the Save or the
// Reset button under the box of a setting presses it.
//
// Parameters: w and h are the size of the pane; x and y are the cell; double
// is true for a double click.
// Returns: the tab, and what the click asks the shell to do.
func (m Model) ProjectClickAt(w, h, x, y int, double bool) (Model, Result) {
	if m.project == nil {
		return m, Result{}
	}
	m.section, m.onItems = Project, true
	plain := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	p := m.rightRows(plain, w, h, -1, -1)
	if y == p.btnRow {
		return m.pressButton(p.btns, x)
	}
	return m.clickItem(p.rowItem, y, double)
}

// ProjectKey gives a key to the Project section that ProjectPane draws: the
// keyboard stays in its rows, because that pane has no list of sections.
//
// Parameters: msg is the key.
// Returns: the tab, and what the key asks the shell to do.
func (m Model) ProjectKey(msg tea.KeyMsg) (Model, Result) {
	if m.project == nil {
		return m, Result{}
	}
	m.section, m.onItems = Project, true
	m, res := m.Key(msg)
	m.onItems = true
	return m, res
}

// CloseProject forgets the open project: the Settings tab lists its own
// sections again, because the settings of a project live in the detail of
// the Projects tab. A tab on the Project section goes to the first section.
//
// Returns: the tab without the project.
func (m Model) CloseProject() Model {
	if m.project == nil {
		return m
	}
	m.project, m.proj, m.projectLoading = nil, ProjectView{}, false
	if m.section == Project {
		m.section, m.ed = Sections[0], nil
	}
	return m
}

// OpenProject opens the Project section on one project: the values its
// manifest owns, read-only, and the settings of [projects.<name>], changed
// with the same editors as the Workspace section. The settings are read in
// the background; the section says so at once.
//
// Parameters: p is the project, as the Projects tab knows it.
// Returns: the tab on the section, and the work that reads the settings.
func (m Model) OpenProject(p ProjectRef) (Model, tea.Cmd) {
	m.project, m.proj = &p, ProjectView{Name: p.Name}
	m.section, m.onItems, m.ed, m.note = Project, true, nil, ""
	m.cursor[Project] = 0
	load := m.b.LoadProject
	if load == nil {
		m.proj.Err = errUnavailable()
		return m, nil
	}
	m.projectLoading = true
	return m, func() tea.Msg { return ProjectLoaded{View: load(p)} }
}

// ProjectOpen is the name of the project the Project section shows; "" when
// none is open.
func (m Model) ProjectOpen() string {
	if m.project == nil {
		return ""
	}
	return m.project.Name
}

// projectTable is the table of smith.toml the open project's settings sit in.
func (m Model) projectTable() string {
	if m.project == nil {
		return ""
	}
	return "[projects." + m.project.Name + "]"
}

// titleOf is the name of a section in the list: the Project section names
// its project.
func (m Model) titleOf(s Section) string {
	if s == Project && m.project != nil {
		return words.Say("settingsview.section.project.named", "name", m.project.Name)
	}
	return s.Title()
}

// projectLoaded takes the settings of the open project; an answer for a
// project that is no longer open is dropped.
func (m Model) projectLoaded(v ProjectView) Model {
	if m.project == nil || v.Name != m.project.Name {
		return m
	}
	m.proj, m.projectLoading = v, false
	return m.clampCursors()
}

// factNote is what Enter on a read-only value says: where to change it.
func (m Model) factNote(f Fact) string {
	return words.Say("settingsview.fact.readonly", "key", f.Key, "file", f.Source)
}

// freshProject reads the open project's settings again after a change, for
// Applied; nil when no project is open.
func (m Model) freshProject() func() *ProjectView {
	load, p := m.b.LoadProject, m.project
	if load == nil || p == nil {
		return func() *ProjectView { return nil }
	}
	ref := *p
	return func() *ProjectView {
		v := load(ref)
		return &v
	}
}
