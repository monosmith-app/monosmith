package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runhistory"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// navParts are the parts of the header of a run view.
func (s Shell) navParts() []runhistory.NavPart {
	return runhistory.NavParts(!s.running(), s.runAt, len(s.runs))
}

// navLine draws the header of a run view in w cells.
func (s Shell) navLine(w int) string {
	var b strings.Builder
	b.WriteString(" ")
	for _, p := range s.navParts() {
		switch {
		case p.Action == "":
			b.WriteString(s.th.Muted(p.Text))
		case p.On:
			b.WriteString(s.th.Accent(p.Text))
		default:
			b.WriteString(s.th.Faint(p.Text))
		}
	}
	return pane.Fit(b.String(), w)
}

// navAction is the action of the header part under column x ("older"
// or "newer"), or "" when x is on plain text.
func (s Shell) navAction(x int) string { return runhistory.ActionAt(s.navParts(), x) }

// openRecent opens the run chosen in the list of recent runs in the right panel
// of Projects. A run in progress keeps the panel: only it can be opened then.
func (s *Shell) openRecent() tea.Cmd {
	at := len(s.runs) - 1 - s.recentAt
	if s.running() && at != len(s.runs)-1 {
		s.say(words.Say("tui.run_nav.a_run_is_in_progress_the"))
		return nil
	}
	s.showRun(at)
	return s.revealRun()
}

// revealRun shows the run panel: Projects opens with the panel beside the list,
// or over it on a narrow screen (where Esc goes back). The keyboard moves to
// the run.
func (s *Shell) revealRun() tea.Cmd {
	cmd := s.gotoTab(commands.TabProjects)
	s.runOpen, s.sideFocus = true, true
	if s.lay().Mode == layout.Narrow {
		s.detailOpen, s.detailCursor = true, 0
	}
	s.keepRunVisible()
	return cmd
}

// closeRun gives the right panel back to the detail of the project. The run
// goes on, and the list of recent runs opens it again.
func (s *Shell) closeRun() {
	s.runOpen, s.detailOpen, s.zoom, s.runFull, s.detailTop, s.sideFocus = false, false, false, false, 0, false
}
