package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
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
		s.say("A run is in progress (" + commands.ShortcutOf("run.stop") + " stops it)")
	case next < 0:
		s.say("This is the oldest run kept")
	case next >= len(s.runs):
		s.say("This is the latest run")
	default:
		s.showRun(next)
	}
}

// repeatArgv is the command the repeat key runs: the run shown, else the last.
func (s Shell) repeatArgv() []string {
	if s.run != nil {
		return s.run.Command
	}
	return s.lastArg
}

// failedArgv is the command line that runs again the tasks of the shown run
// that failed, were interrupted or never started; nil when there are none.
// Tasks of different names on different projects run as a grid of both, the
// ones that already passed coming back from the cache.
func (s Shell) failedArgv() []string {
	if s.run == nil || !s.run.Finished {
		return nil
	}
	var tasks, projects []string
	add := func(list []string, v string) []string {
		for _, have := range list {
			if have == v {
				return list
			}
		}
		return append(list, v)
	}
	for _, t := range s.run.Ordered() {
		if t.Status != runview.Failed && t.Status != runview.Skipped {
			continue
		}
		project, task, ok := strings.Cut(t.ID, ":")
		if !ok {
			continue
		}
		projects, tasks = add(projects, project), add(tasks, task)
	}
	if len(tasks) == 0 {
		return nil
	}
	return append(append([]string{"run"}, tasks...), "-p", strings.Join(projects, ","))
}

// copyLog puts the log shown on the clipboard: one task's output, or all of it.
func (s *Shell) copyLog() {
	lines := s.logLines()
	if len(lines) == 0 {
		s.say("Nothing to copy: the log is empty")
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
		const none = "nothing has run yet"
		for _, id := range []string{"run.follow", "run.scroll", "run.again-failed", "run.copy-log", "run.older", "run.newer"} {
			r[id] = none
		}
		if s.lastArg == nil {
			r["run.again"] = none
		}
		return r
	}
	if s.failedArgv() == nil {
		r["run.again-failed"] = "no task failed or was skipped"
	}
	if s.runAt <= 0 {
		r["run.older"] = "this is the oldest run kept"
	}
	if s.runAt >= len(s.runs)-1 {
		r["run.newer"] = "this is the latest run"
	}
	if len(s.logLines()) == 0 {
		r["run.copy-log"] = "the log is empty"
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
		return fmt.Sprintf("■ stopped after %s · %s", took, line)
	case !last.OK():
		return fmt.Sprintf("✖ failed (code %d) in %s · %s", last.Exit, took, line)
	}
	return fmt.Sprintf("✔ done in %s · %s", took, line)
}

// runActions are the actions of the Run tab's detail: only those that can be
// done now are offered, as the footer does for keys.
func (s Shell) runActions() []detail.Action {
	var out []detail.Action
	if !s.running() && s.repeatArgv() != nil {
		out = append(out, detail.Action{ID: "x:rerun", Label: "Run again"})
	}
	if !s.running() && s.failedArgv() != nil {
		out = append(out, detail.Action{ID: "x:rerun-failed", Label: "Run the unfinished tasks again"})
	}
	if len(s.logLines()) > 0 {
		out = append(out, detail.Action{ID: "x:copy-log", Label: "Copy the log"})
	}
	return append(out, detail.Action{ID: "x:projects", Label: "Show the projects"})
}
