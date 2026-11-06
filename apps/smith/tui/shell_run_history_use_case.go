package tui

import (
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runhistory"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The Run tab keeps the runs of the session (the newest last), so an earlier
// result can be read again after another run started. The run shown is s.run;
// while one is going it is the one shown, and the history keys wait for it.

// maxRunsKept bounds the runs kept in memory: each holds its captured output.
const maxRunsKept = 20

// remember adds a run that just started to the history and shows it.
func (s *Shell) remember(r *runview.Run) {
	s.runs = append(s.runs, r)
	if len(s.runs) > maxRunsKept {
		s.runs = append([]*runview.Run(nil), s.runs[len(s.runs)-maxRunsKept:]...)
	}
	s.runAt = len(s.runs) - 1
}

// showRun shows another run of the history: the first failed task is selected,
// so its output is the first thing read.
func (s *Shell) showRun(at int) {
	if at < 0 || at >= len(s.runs) || s.running() {
		return
	}
	s.runAt, s.run, s.runHome = at, s.runs[at], false
	s.runSel, s.runTop, s.log = s.firstFailed(), 0, logview.Log{}
	s.keepRunVisible()
}

// firstFailed is the ID of the first task of the shown run that failed, or "".
// A task that failed on its own comes before one that was only interrupted when
// the run ended, because the first has the output that says why.
func (s *Shell) firstFailed() string {
	if s.run == nil {
		return ""
	}
	interrupted := ""
	for _, t := range s.run.Ordered() {
		switch {
		case t.Status != runview.Failed:
		case t.Reason == "interrupted":
			if interrupted == "" {
				interrupted = t.ID
			}
		default:
			return t.ID
		}
	}
	return interrupted
}

// stepRun moves through the history: dir -1 is older, 1 is newer.
func (s *Shell) stepRun(dir int) {
	next := s.runAt + dir
	switch {
	case s.running():
		s.say(words.Say("tui.run_history.a_run_is_in_progress_stops", "shortcutOf", commands.ShortcutOf("run.stop")))
	case next < 0:
		s.say(words.Say("tui.run_history.this_is_the_oldest_run_kept"))
	case next >= len(s.runs):
		s.say(words.Say("tui.run_history.this_is_the_latest_run"))
	default:
		s.showRun(next)
	}
}

// repeatArgv is the command the repeat key runs: the run shown, else the last.
// A command that makes something that exists afterwards is not offered again:
// running `smith add` twice fails the second time, because the project is there.
func (s Shell) repeatArgv() []string {
	argv := s.lastArg
	if s.run != nil {
		argv = s.run.Command
	}
	if !repeatable(argv) {
		return nil
	}
	return argv
}

// failedArgv is the command line that runs again the tasks of the shown run
// that failed, were interrupted or never started; nil when there are none.
// Tasks of different names on different projects run as a grid of both, the
// ones that already passed coming back from the cache.
func (s Shell) failedArgv() []string {
	if s.run == nil || !s.run.Finished {
		return nil
	}
	var ids []string
	for _, t := range s.run.Ordered() {
		if t.Status == runview.Failed || t.Status == runview.Skipped {
			ids = append(ids, t.ID)
		}
	}
	return runhistory.RerunFailed(ids)
}

// copyLog puts the log shown on the clipboard: one task's output, or all of it.
func (s *Shell) copyLog() {
	lines := s.logLines()
	if len(lines) == 0 {
		s.say(words.Say("tui.run_history.nothing_to_copy_the_log_is"))
		return
	}
	plain := make([]string, len(lines))
	for i, l := range lines {
		plain[i] = pane.Strip(l)
	}
	s.copyText(strings.Join(plain, "\n"))
}

// runReasons are why the keys of the Run tab cannot work now ("" keys are fine).
func (s Shell) runReasons() map[string]string {
	r := map[string]string{}
	if s.run == nil {
		none := words.Say("tui.run_history.none")
		for _, id := range []string{"run.follow", "run.scroll", "run.again-failed", "run.copy-log", "run.older", "run.newer", "run.open-recent", "run.all-output"} {
			r[id] = none
		}
		if s.lastArg == nil {
			r["run.again"] = none
		}
		return r
	}
	if s.repeatArgv() == nil {
		r["run.again"] = notRepeatableReason(s.run.Command)
	}
	if s.failedArgv() == nil {
		r["run.again-failed"] = words.Say("tui.run_history.no_task_failed_or_was_skipped")
	}
	if s.runAt <= 0 {
		r["run.older"] = words.Say("tui.run_history.this_is_the_oldest_run_kept_2")
	}
	if s.runAt >= len(s.runs)-1 {
		r["run.newer"] = words.Say("tui.run_history.this_is_the_latest_run_2")
	}
	if len(s.logLines()) == 0 {
		r["run.copy-log"] = words.Say("tui.run_history.the_log_is_empty")
	}
	if s.runHome {
		// The start screen lists the runs: the keys of a run's own view wait.
		home := words.Say("tui.run_history.home")
		for _, id := range []string{"run.follow", "run.again-failed", "run.copy-log", "run.older", "run.newer", "run.all-output"} {
			r[id] = home
		}
	} else {
		view := words.Say("tui.run_history.view")
		r["run.open-recent"] = view
	}
	return r
}

// ExitSummary is the line printed in the scrollback when the interface ends: how
// the last run went and the command that repeats it. It is empty when nothing ran.
func (s Shell) ExitSummary() string {
	if s.exitNote != "" {
		return s.exitNote
	}
	if len(s.runs) == 0 {
		return ""
	}
	last := s.runs[len(s.runs)-1]
	took := last.RunTime(s.now()).Round(100 * time.Millisecond)
	line := last.CommandLine()
	switch {
	case !last.Finished || last.Cancelled:
		return words.Say("tui.run_history.stopped_after", "took", took, "line", line)
	case !last.OK():
		return words.Say("tui.run_history.failed_code_in", "exit", last.Exit, "took", took, "line", line)
	}
	return words.Say("tui.run_history.done_in", "took", took, "line", line)
}

// runActions are the actions of the Run tab's detail: only those that can be
// done now are offered, as the footer does for keys.
func (s Shell) runActions() []detail.Action {
	var out []detail.Action
	if !s.running() && s.repeatArgv() != nil {
		out = append(out, detail.Action{ID: "x:rerun", Label: words.Say("tui.run_history.run_again")})
	}
	if !s.running() && s.failedArgv() != nil {
		out = append(out, detail.Action{ID: "x:rerun-failed", Label: words.Say("tui.run_history.run_the_unfinished_tasks_again")})
	}
	if len(s.logLines()) > 0 {
		out = append(out, detail.Action{ID: "x:copy-log", Label: words.Say("tui.run_history.copy_the_log")})
	}
	return append(out, detail.Action{ID: "x:projects", Label: words.Say("tui.run_history.show_the_projects")})
}
