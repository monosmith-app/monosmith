package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectlist"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The projects pinned to the rail: the person's own, kept outside the workspace.

// pinned is true when the project is pinned to the rail.
func (s Shell) pinned(name string) bool { return name != "" && slices.Contains(s.pins, name) }

// togglePin pins the project to the rail, or unpins it when it is pinned
// already, and keeps the pins in the person's preferences. It says what it did.
func (s *Shell) togglePin(name string) {
	if name == "" {
		s.say(words.Say("tui.projects.reason.no_project"))
		return
	}
	if i := slices.Index(s.pins, name); i >= 0 {
		s.pins = slices.Delete(slices.Clone(s.pins), i, i+1)
		s.say(words.Say("tui.pins.unpinned", "project", name))
	} else {
		s.pins = append(slices.Clone(s.pins), name)
		s.say(words.Say("tui.pins.pinned", "project", name))
	}
	if s.savePins != nil {
		s.savePins(slices.Clone(s.pins))
	}
	s.refreshPalette()
}

// openPinned opens a pinned project: the Projects view, with the project
// selected and the keyboard on the list. The filter goes when it hides the
// project.
func (s *Shell) openPinned(name string) tea.Cmd {
	if _, ok := s.list.Project(name); !ok {
		s.say(words.Say("tui.pins.missing", "project", name))
		return nil
	}
	cmd := s.gotoTab(commands.TabProjects)
	listed, _ := s.list.Listed()
	if !slices.ContainsFunc(listed, func(p projectlist.Project) bool { return p.Name == name }) {
		s.list = s.list.WithFilter("")
	}
	s.list = s.list.SelectProject(name)
	s.sideFocus = false
	return cmd
}
