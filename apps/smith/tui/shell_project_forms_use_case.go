package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

// openProjectForm opens the form titled title with its project question, the
// step with the key, as a pick list of names that starts on the selected
// project. With no names the question stays free text.
func (s *Shell) openProjectForm(title, key string, names []string) tea.Cmd {
	g := s.guide
	if base, ok := g.ActionTitled(title); ok {
		g = g.WithAction(guide.WithProjectPick(base, key, names))
	}
	if g, ok := g.WithDefault(key, s.selectedProject()).Pick(title); ok {
		s.guide = g
	}
	return s.gotoTab(commands.TabGuide)
}

// openPresentedForm opens the form of a catalogue action (`smith kind`,
// `smith tauri add`), built from the command's description, on the selected
// project. Its project question is the argument named project; names are the
// projects it lists. A form the menu does not have opens nothing.
func (s *Shell) openPresentedForm(id string, names []string) tea.Cmd {
	a, ok := s.guide.ActionPresenting(id)
	if !ok {
		return nil
	}
	return s.openProjectForm(a.Title, "arg:project", names)
}

// projectsWhere are the names of the projects that keep holds for.
func (s Shell) projectsWhere(keep func(projectlist.Project) bool) []string {
	var names []string
	for _, p := range s.list.Projects() {
		if keep(p) {
			names = append(names, p.Name)
		}
	}
	return names
}

// canAddTauri reports whether `smith tauri add` applies to a project: a Node
// app that is not a Tauri app yet.
func canAddTauri(p projectlist.Project) bool { return isWebApp(p) && !isTauriApp(p) }
