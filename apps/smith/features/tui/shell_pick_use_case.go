package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
)

// The projects the person picks with Space in the list of Projects.

// togglePick picks the project under the cursor, or lets it go when it is
// picked already. Picking a project moves the cursor to the next row, so
// Space, Space picks two in a row.
func (s *Shell) togglePick() {
	p, ok := s.list.Selected()
	if !ok {
		return
	}
	if s.picked == nil {
		s.picked = map[string]bool{}
	}
	if s.picked[p.Name] {
		delete(s.picked, p.Name)
		return
	}
	s.picked[p.Name] = true
	s.list = s.list.Move(1)
}

// pickedNames are the picked projects that are still in the workspace, in the
// order of the list. They count only on Projects: another view runs on the
// selected project, as before.
func (s Shell) pickedNames() []string {
	if len(s.picked) == 0 || s.tab != commands.TabProjects {
		return nil
	}
	var names []string
	for _, p := range s.list.Projects() {
		if s.picked[p.Name] {
			names = append(names, p.Name)
		}
	}
	return names
}
