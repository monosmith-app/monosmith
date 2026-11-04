package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

// installDependency opens the form that installs packages from the Deps view.
// Every project is offered with its language, so a project with several parts
// shows which one receives the package. The projects that declare the package
// under the cursor start ticked, the workspace root included when the row is a
// (root) one; the person still confirms them.
func (s *Shell) installDependency() tea.Cmd {
	var offered []guide.InstallTarget
	hasNode := false
	for _, p := range s.list.Projects() {
		hasNode = hasNode || p.Language == "node"
		label := p.Name
		if p.Language != "" {
			label += " (" + p.Language + ", " + p.Root + ")"
		}
		offered = append(offered, guide.InstallTarget{Label: label, Value: p.Name})
	}
	if hasNode {
		// Only the Node root package.json takes packages today; the other languages show no root.
		offered = append([]guide.InstallTarget{guide.RootInstallTarget}, offered...)
	}
	if len(offered) == 0 {
		s.say("The workspace has no project to install into")
		return nil
	}
	var preset []string
	if _, targets, ok := s.deps.RemoveTargets(); ok {
		for _, t := range targets {
			if t.Value != "." || hasNode {
				preset = append(preset, t.Value)
			}
		}
	}
	g := s.guide.WithAction(guide.InstallPackageAction(offered, preset))
	g, picked := g.Pick(guide.InstallPackageTitle)
	if !picked {
		return nil
	}
	s.guide = g
	return s.gotoTab(commands.TabGuide)
}
