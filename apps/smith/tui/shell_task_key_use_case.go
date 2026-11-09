package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// taskKey runs the task of the selected project that the key name belongs to: a
// task the keymap has no key for shows a button with the key chosen for it (see
// chooseTaskKeys), and this is where that key works. It reports whether the key
// was one of them, so any other key goes on.
func (s *Shell) taskKey(name string) (tea.Cmd, bool) {
	if s.tab != commands.TabProjects || !s.inWorkspace || s.editing() || s.formOpen() {
		return nil, false
	}
	p, ok := s.list.Selected()
	if !ok {
		return nil, false
	}
	for task, key := range s.taskKeys(p) {
		if key != name {
			continue
		}
		run, _ := commands.ByID("run.build")
		if reason := run.Unavailable(s.state()); reason != "" {
			s.say(words.Say("tui.keys.is_unavailable", "title", taskLabel(task, p.Kind), "reason", reason))
			return nil, true
		}
		return s.activate("x:run:" + task), true
	}
	return nil, false
}
