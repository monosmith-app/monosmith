package settingsview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

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
