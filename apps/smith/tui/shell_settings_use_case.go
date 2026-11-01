package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// The Settings tab (package settingsview) changes plugins, the style and the
// settings of smith.toml in this process, with the functions the matching
// smith commands call, so nothing leaves the interface. These are the few
// places the shell hands it keys, clicks and messages, and takes its results.

// settingsForms are the guided forms the tab's last section opens.
func settingsForms(actions []guide.Action) []settingsview.Form {
	forms := make([]settingsview.Form, 0, len(actions))
	for _, a := range actions {
		forms = append(forms, settingsview.Form{Title: a.Title, Hint: a.Hint})
	}
	return forms
}

// settingsKey gives the open Settings tab a key.
func (s *Shell) settingsKey(msg tea.KeyMsg) tea.Cmd {
	if why := s.settingsRefusal(msg.String()); why != "" {
		s.say(why)
		return nil
	}
	var res settingsview.Result
	s.settings, res = s.settings.Key(msg)
	return s.settingsResult(res)
}

// settingsRefusal is why a key of the tab does nothing on the row now, in the
// words every refused key uses; "" when the key works or is not one of those.
func (s Shell) settingsRefusal(key string) string {
	if s.settings.Modal() {
		return ""
	}
	reasons := s.settings.Reasons()
	for _, id := range []string{"settings.change", "settings.trust", "settings.forget", "settings.unset"} {
		if c, ok := commands.ByID(id); ok && slices.Contains(c.Keys, key) && reasons[id] != "" {
			return c.Title + " is unavailable: " + reasons[id]
		}
	}
	return ""
}

// settingsClick gives the tab a click at (x, y) of its body.
func (s *Shell) settingsClick(w, h, x, y int, double bool) tea.Cmd {
	var res settingsview.Result
	s.settings, res = s.settings.Click(w, h, x, y, double)
	return s.settingsResult(res)
}

// settingsMsg takes what the tab's background work reports.
func (s *Shell) settingsMsg(msg settingsview.Msg) tea.Cmd {
	var res settingsview.Result
	s.settings, res = s.settings.Handle(msg)
	return s.settingsResult(res)
}

// settingsResult does what the tab asked of the shell: a toast, another
// theme, a guided form, background work.
func (s *Shell) settingsResult(res settingsview.Result) tea.Cmd {
	if res.Note != "" {
		level := toast.Success
		if res.Failed {
			level = toast.Error
		}
		s.notify(res.Note, level)
	}
	if res.Theme != nil {
		s.th = theme.New(*res.Theme)
	}
	if res.Run != nil {
		return s.execute(res.Run, false)
	}
	if res.Search {
		return s.openPluginSearch(res.Term)
	}
	if res.Form != "" {
		return s.perform(commands.Command{Kind: commands.Guided, Guide: res.Form})
	}
	return res.Cmd
}

// settingsSectionFor is the section of the Settings tab that shows what a
// read-only command line prints: smith plugin, smith recipe, smith config list.
func settingsSectionFor(argv []string) (settingsview.Section, bool) {
	if len(argv) == 0 {
		return 0, false
	}
	rest := argv[1:]
	switch argv[0] {
	case "plugin":
		if len(rest) == 0 || (rest[0] == "list" && len(rest) == 1) {
			return settingsview.Plugins, true
		}
	case "architecture":
		if len(rest) == 0 {
			return settingsview.Architecture, true
		}
	case "recipe":
		if len(rest) == 0 {
			return settingsview.Recipes, true
		}
	case "config":
		if len(rest) == 0 || (rest[0] == "list" && len(rest) == 1) {
			return settingsview.Workspace, true
		}
	}
	return 0, false
}

// showSettings opens the Settings tab on a section.
func (s *Shell) showSettings(section settingsview.Section) tea.Cmd {
	cmd := s.gotoTab(commands.TabSettings)
	s.settings = s.settings.Show(section)
	return cmd
}
