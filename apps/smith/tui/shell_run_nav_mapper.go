package tui

import (
	"strings"

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

// navAction is the action of the header part under column x ("start", "older"
// or "newer"), or "" when x is on plain text.
func (s Shell) navAction(x int) string { return runhistory.ActionAt(s.navParts(), x) }

// lastRun is the newest run kept, or the one shown when none is kept.
func (s Shell) lastRun() *runview.Run {
	if len(s.runs) == 0 {
		return s.run
	}
	return s.runs[len(s.runs)-1]
}

// openStart shows the start screen of the Run tab, with the shown run chosen in
// its list of recent runs. A run in progress keeps the screen.
func (s *Shell) openStart() {
	if s.running() {
		s.say(words.Say("tui.run_nav.a_run_is_in_progress_the"))
		return
	}
	s.runSel, s.runHome = "", true
	s.recentAt = len(s.runs) - 1 - s.runAt
}

// openRecent shows the run chosen in the list of recent runs.
func (s *Shell) openRecent() {
	s.showRun(len(s.runs) - 1 - s.recentAt)
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
	pointer := " "
	if i == s.recentAt {
		pointer = s.th.Icons.Pointer
	}
	return runhistory.Line(runhistory.Row{
		Pointer: pointer, Command: r.CommandLine(), Result: s.runResult(r),
		Took: shortDuration(r.RunTime(now)), Ago: ago, Selected: i == s.recentAt,
	}, w, s.th.Muted, s.th.Selected)
}

// runResult is how a run stands, as a mark and a word: running, passed, failed
// or stopped.
func (s Shell) runResult(r *runview.Run) string {
	switch {
	case !r.Finished:
		return s.th.Info(words.Say("tui.run_nav.running", "spin", s.th.Spin(s.tick)))
	case r.Cancelled:
		return s.th.Warning(words.Say("tui.run_nav.stopped", "skipped", s.th.Icons.Skipped))
	case r.OK():
		return s.th.Success(words.Say("tui.run_nav.passed", "passed", s.th.Icons.Passed))
	}
	return s.th.Failure(words.Say("tui.run_nav.failed", "failed", s.th.Icons.Failed))
}
