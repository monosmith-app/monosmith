package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
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

// openRecent opens the run chosen in a list of recent runs in the Runs view. A
// run in progress keeps the view: only it can be opened then.
func (s *Shell) openRecent() tea.Cmd {
	at := len(s.runs) - 1 - s.recentAt
	if s.running() && at != len(s.runs)-1 {
		s.say(words.Say("tui.run_nav.a_run_is_in_progress_the"))
		return nil
	}
	s.showRun(at)
	return s.revealRun()
}

// revealRun opens the Runs view on the run shown, from whichever view it was
// started or opened; Esc goes back to that view. A narrow screen shows the tasks.
func (s *Shell) revealRun() tea.Cmd {
	if s.tab != commands.TabRuns {
		s.runFrom = s.tab
	}
	cmd := s.gotoTab(commands.TabRuns)
	s.runOnLog = false
	s.keepRunVisible()
	return cmd
}

// closeRun leaves the Runs view for the view the run was opened from (Home when
// there is none). The run goes on, and key 3 opens it again.
func (s *Shell) closeRun() tea.Cmd {
	to := s.runFrom
	if to == "" || to == commands.TabRuns || !s.canShow(to) {
		to = commands.TabHome
	}
	s.runFrom = ""
	return s.gotoTab(to)
}
