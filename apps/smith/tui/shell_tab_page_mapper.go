package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
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
		return fmt.Sprintf("Projects (%d)", s.list.Len())
	case commands.TabRun:
		if s.run != nil {
			return s.run.CommandLine()
		}
	case commands.TabGuide:
		if t := s.guide.Title(); t != "" {
			return t
		}
	}
	return tabTitles[s.tab]
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
