package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// navPart is one part of the run header: its text, the action a click on it
// stands for ("" for plain text) and whether that action can be done now.
type navPart struct {
	text, action string
	on           bool
}

// navParts are the parts of the header of a run view: "‹ Start · run 2 of 5 ·
// ‹ older  newer ›". A part that cannot be used now is drawn faint; a click on
// it still says why nothing happened.
func (s Shell) navParts() []navPart {
	idle := !s.running()
	return []navPart{
		{words.Say("tui.run_nav.start"), "start", idle},
		{" · ", "", false},
		{words.Say("tui.run_nav.run_of", "value", s.runAt+1, "len", len(s.runs)), "", false},
		{" · ", "", false},
		{words.Say("tui.run_nav.older"), "older", idle && s.runAt > 0},
		{"  ", "", false},
		{words.Say("tui.run_nav.newer"), "newer", idle && s.runAt < len(s.runs)-1},
	}
}

// navLine draws the header of a run view in w cells.
func (s Shell) navLine(w int) string {
	var b strings.Builder
	b.WriteString(" ")
	for _, p := range s.navParts() {
		switch {
		case p.action == "":
			b.WriteString(s.th.Muted(p.text))
		case p.on:
			b.WriteString(s.th.Accent(p.text))
		default:
			b.WriteString(s.th.Faint(p.text))
		}
	}
	return pane.Fit(b.String(), w)
}

// navAction is the action of the header part under column x ("start", "older"
// or "newer"), or "" when x is on plain text.
func (s Shell) navAction(x int) string {
	at := 1
	for _, p := range s.navParts() {
		w := pane.Width(p.text)
		if x >= at && x < at+w {
			return p.action
		}
		at += w
	}
	return ""
}

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
	rows = min(rows, len(s.runs))
	first = max(s.recentAt-rows+1, 0)
	for i := first; i < first+rows; i++ {
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
		ago = agoText(now.Sub(r.Ended))
	}
	cmdW := max(w-3-1-10-1-8-1-10, 12)
	pointer := " "
	if i == s.recentAt {
		pointer = s.th.Icons.Pointer
	}
	line := " " + pointer + " " + pad(r.CommandLine(), cmdW) + " " + pad(s.runResult(r), 10) + " " +
		fmt.Sprintf("%8s", shortDuration(r.RunTime(now))) + " " + s.th.Muted(fmt.Sprintf("%10s", ago))
	if i == s.recentAt {
		return s.th.Selected(pane.Fit(pane.Strip(line), w))
	}
	return pane.Fit(line, w)
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
