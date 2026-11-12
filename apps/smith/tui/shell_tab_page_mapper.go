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

// Title implements Page. The active tab is lit in the tab strip, so the frame has
// no title that only says which tab is open (a fact shows once). Only what the
// tab strip does not say is a title: the form open on the Start tab, or what the
// Start tab does while a box covers its form.
func (p tabPage) Title() string {
	s := p.sh
	if s.tab != commands.TabGuide {
		return ""
	}
	if t := s.guide.Title(); t != "" {
		return t
	}
	return s.guideModalName()
}

// guideModalName is what the Start tab is doing while a box is open over it and
// the form is no longer on screen: the form that waits for its confirmation, the
// composer, the adopt wizard. It is "" when no such box is open, and the tab
// has no title.
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

// Footer implements Page. On a button above the rule of a split detail, Enter
// goes to the run button and does not run yet: its hint says so.
func (p tabPage) Footer() (view, tail []commands.Hint) {
	view, tail = commands.Hints(p.sh.state())
	s := p.sh
	if s.sideKeys() && !s.detailOnRun && !s.runShown() && s.detailPanel().Splits() {
		for i, h := range view {
			if h.Key == "enter" && h.Label == words.Say("commands.side.open.footer") {
				view[i].Label = words.Say("commands.side.open.footer.chooser")
			}
		}
	}
	return view, tail
}

// Update implements Page: keys go to the view's own handler.
func (p tabPage) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok {
		return p.sh.paneKey(key)
	}
	return nil
}

// View implements Page.
func (p tabPage) View(width, height int) content { return p.sh.mainContent(width, height) }
