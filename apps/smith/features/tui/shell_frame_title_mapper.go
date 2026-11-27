package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// mainTitle is the title in the top border of the main frame: the open view's
// name, or "‹ View / item" while the detail fills the frame (opened with Enter
// on a narrow screen, or zoomed on a wide one). Esc goes back to the list.
func (s Shell) mainTitle() string {
	view := s.viewName(s.tab)
	if s.tab == commands.TabProjects && s.list.Filter() != "" {
		// A filter lists some of the projects: the title says how many of all.
		listed, _ := s.list.Listed()
		view = words.Say("tui.projects.title_filtered", "view", view, "shown", len(listed), "total", s.list.Len())
	}
	if s.runShown() {
		// The Runs view names the run: "Run · smith test", or on a narrow screen
		// "‹ Runs / smith test", the frame of an item opened in its place.
		if s.lay().Mode == layout.Narrow && !s.fullRun() {
			return words.Say("tui.frame.back", "view", view, "item", s.run.CommandLine())
		}
		return words.Say("tui.frame.run", "command", s.run.CommandLine())
	}
	if s.tab == commands.TabSettings && !s.lay().HasDetail() && !s.settings.OnSections() {
		// On a narrow screen Enter opened the section in the frame; Esc goes back.
		return words.Say("tui.frame.back", "view", view, "item", s.settings.Section().Title())
	}
	if s.detailFull() {
		if item := s.detailTitle(); item != "" {
			return words.Say("tui.frame.back", "view", view, "item", item)
		}
	}
	return view
}

// detailTitle is the title of the detail's frame: what the detail tells about.
// It is the selected project, the selected check, the selected package (or the
// one-version policy), the selected release or the settings section; "" for a
// view without a detail.
func (s Shell) detailTitle() string {
	switch s.tab {
	case commands.TabProjects:
		if p, ok := s.list.Selected(); ok && s.projMenu {
			return words.Say("tui.projects.menu_title", "name", p.Name)
		} else if ok {
			return p.Name
		}
		return words.Say("tui.detail.workspace")
	case commands.TabDoctor:
		if checks := s.shownFixChecks(); len(checks) > 0 {
			// The detail shows a fix: it is titled by the check the fix is about.
			if f, ok := s.fixFinding(checks); ok {
				return f.Check
			}
		}
		if f, ok := s.doctor.Selected(); ok {
			return f.Check
		}
		if e, ok := s.doctor.SelectedExtra(); ok {
			return e.Title
		}
		return words.Say("tui.detail.workspace")
	case commands.TabDeps:
		return s.deps.DetailTitle()
	case commands.TabRelease:
		return s.release.DetailTitle()
	case commands.TabSettings:
		return s.settings.PaneTitle()
	}
	return ""
}

// mainFocused is true while the main frame has the keyboard: always, unless
// the detail beside it has it.
func (s Shell) mainFocused() bool {
	if s.tab == commands.TabSettings && s.lay().HasDetail() {
		// The keyboard is in the list of sections or in the open section beside it.
		return s.settings.OnSections()
	}
	return s.detailFull() || !s.sideKeys()
}
