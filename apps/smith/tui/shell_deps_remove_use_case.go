package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

// removeDependency removes the package under the cursor of the Deps view from
// the projects that declare it. A package declared in one project goes straight
// to the confirmation with that project as -w. When several declare it, a
// multiple choice of projects comes first: smith never picks for the person.
func (s *Shell) removeDependency() tea.Cmd {
	name, targets, ok := s.deps.RemoveTargets()
	if !ok {
		s.say("No dependency is selected")
		return nil
	}
	if len(targets) == 1 {
		return s.execute([]string{"uninstall", "-w", targets[0].Value, name}, false)
	}
	var offered []guide.RemoveTarget
	for _, t := range targets {
		offered = append(offered, guide.RemoveTarget{Label: t.Label, Value: t.Value})
	}
	title := guide.RemovePackageTitle(name)
	g := s.guide.WithAction(guide.RemovePackageAction(name, offered))
	g, picked := g.Pick(title)
	if !picked {
		return nil
	}
	s.guide = g
	return s.gotoTab(commands.TabGuide)
}
