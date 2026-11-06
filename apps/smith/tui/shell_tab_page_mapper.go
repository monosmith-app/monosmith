package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// tabPage is the open tab seen as a Page. The existing view models stay as
// they are; the page adapts them to the shell's one pane.
type tabPage struct{ sh *Shell }

// page is the Page of the open tab.
func (s *Shell) page() Page { return tabPage{sh: s} }

// Title implements Page.
func (p tabPage) Title() string {
	s := p.sh
	switch s.tab {
	case commands.TabProjects:
		if s.detailFull() {
			return s.list.SelectedName()
		}
		return words.Say("tui.tab_page.projects", "len", s.list.Len())
	case commands.TabRun:
		if s.run != nil {
			return s.run.CommandLine()
		}
	case commands.TabGuide:
		if t := s.guide.Title(); t != "" {
			return t
		}
		if t := s.guideModalName(); t != "" {
			return t
		}
	}
	return tabTitles()[s.tab]
}

// guideModalName is what the Start tab is doing while a box is open over it and
// the form is no longer on screen: the form that waits for its confirmation, the
// composer, the adopt wizard. It is "" when no such box is open, and the tab
// keeps its own name.
func (s Shell) guideModalName() string {
	if s.tab != commands.TabGuide {
		return ""
	}
	switch {
	case s.confirm != nil && s.Form != "":
		return s.Form
	case s.compose != nil:
		return words.Say("tui.tab_page.compose_a_stack")
	case s.adopt != nil:
		return words.Say("tui.tab_page.bring_into_monosmith")
	}
	return ""
}

// Footer implements Page.
func (p tabPage) Footer() (view, tail []commands.Hint) { return commands.Hints(p.sh.state()) }

// Update implements Page: keys go to the view's own handler.
func (p tabPage) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok {
		return p.sh.paneKey(key)
	}
	return nil
}

// View implements Page.
func (p tabPage) View(width, height int) content { return p.sh.mainContent(width, height) }
