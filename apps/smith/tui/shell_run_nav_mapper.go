package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runhistory"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
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
	s.runOpen, s.recentFocus, s.sideFocus = true, false, true
	if s.lay().Mode == layout.Narrow {
		s.detailOpen, s.detailCursor = true, 0
	}
	s.keepRunVisible()
	return cmd
}

// closeRun gives the right panel back to the detail of the project. The run
// goes on, and the list of recent runs opens it again.
func (s *Shell) closeRun() {
	s.runOpen, s.detailOpen, s.zoom, s.detailTop, s.sideFocus = false, false, false, 0, false
}

// recentCollapsed is how many runs the block of recent runs shows until it is
// opened.
const recentCollapsed = 3

// recentSection is the block of recent runs at the foot of the projects pane: a
// rule, a title and the newest runs. Closed, it shows the last three and says
// how many more are kept. Opened (the keyboard is on it, or a click chose a
// run) it grows upward to list every run kept, as far as room lines allow, with
// the cursor on the chosen run. It is empty when there is no run or no room.
// The hits are relative to its first line.
func (s Shell) recentSection(w, room int) content {
	c := content{hits: map[int]hit{}}
	n := len(s.runs)
	if !s.recentFocus {
		n = min(n, recentCollapsed)
	}
	n = min(n, room-2)
	if n < 1 {
		return c
	}
	title := s.th.Title(words.Say("tui.main.recent_runs"))
	note := s.th.Muted(words.Say("tui.main.up_down_choose_enter_or_double"))
	switch {
	case s.recentFocus:
		title = s.th.Accent(words.Say("tui.main.recent_runs"))
		note = s.th.Muted(words.Say("tui.main.recent_runs.opened"))
	case len(s.runs) > n:
		note = s.th.Muted(words.Count("tui.main.recent_runs.more", len(s.runs)-n))
	}
	c.lines = append(c.lines, s.th.Border(strings.Repeat("─", w), false), pane.Fit(" "+title+"  "+note, w))
	rows, first := s.recentRows(w, n)
	for i := range rows {
		c.hits[len(c.lines)] = hit{kind: "recent", index: first + i}
		c.lines = append(c.lines, rows[i])
	}
	return c
}

// recentRows are the rows of the list of recent runs that fit in rows lines,
// the chosen one in view; first is the position of the first row in the list.
func (s Shell) recentRows(w, rows int) (lines []string, first int) {
	first, n := runhistory.Window(s.recentAt, len(s.runs), rows)
	for i := first; i < first+n; i++ {
		lines = append(lines, s.recentLine(i, w))
	}
	return lines, first
}

// recentLine is the row of the i-th newest run: its command, how it ended, how
// long it took and how long ago.
func (s Shell) recentLine(i, w int) string {
	r := s.runs[len(s.runs)-1-i]
	now := s.now()
	ago := "now"
	if r.Finished {
		ago = runhistory.Ago(now.Sub(r.Ended))
	}
	chosen := i == s.recentAt && s.recentFocus
	pointer := " "
	if chosen {
		pointer = s.th.Icons.Pointer
	}
	return runhistory.Line(runhistory.Row{
		Pointer: pointer, Command: r.CommandLine(), Result: s.runResult(r),
		Took: shortDuration(r.RunTime(now)), Ago: ago, Selected: chosen,
	}, w, s.th.Muted, s.th.Selected)
}

// runResult is how a run stands, as a mark and a word: running, passed, failed
// or stopped.
func (s Shell) runResult(r *runview.Run) string {
	switch {
	case !r.Finished:
		return s.th.Info(words.Say("tui.run_nav.running", "spin", s.th.Spin(s.tick)))
	case r.Cancelled:
		return s.th.Muted(words.Say("tui.run_nav.stopped", "cancelled", s.th.Icons.Cancelled))
	case r.OK():
		return s.th.Success(words.Say("tui.run_nav.passed", "passed", s.th.Icons.Passed))
	}
	return s.th.Failure(words.Say("tui.run_nav.failed", "failed", s.th.Icons.Failed))
}
