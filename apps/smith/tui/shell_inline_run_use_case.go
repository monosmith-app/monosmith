package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// A task run (smith run, build, test, lint...) leaves the interface: Bubble
// Tea gives the terminal to the command (tea.Exec), which prints into the real
// scrollback with its own live status line, then waits for a key and the
// interface comes back. Nothing is captured, so the output stays after the
// interface ends and can be selected like any terminal text. The process is
// injected through Sources.Inline, so tests run a script instead.

// inlineRun is one finished inline run, as the Run tab lists it.
type inlineRun struct {
	Argv      []string
	Code      int
	Cancelled bool
	Failed    bool
	Took      time.Duration
}

// OK reports whether the run succeeded.
func (r inlineRun) OK() bool { return !r.Cancelled && !r.Failed && r.Code == 0 }

// inlineDoneMsg reports that the terminal came back from an inline run.
type inlineDoneMsg struct {
	cmd *runner.Inline
	err error
}

// cliVerbs are the commands that took the place of the views that left the
// interface (plugins, recipes, settings, the guide) and of the palette's
// read-outs: they print and end, so they use the terminal like a task run.
var cliVerbs = map[string]bool{
	"plugin": true, "recipe": true, "config": true, "info": true, "graph": true, "affected": true,
	"explain": true, "projects": true, "stacks": true, "kinds": true, "commands": true, "help": true,
}

// runsInline reports whether argv is a task run, or a command that prints and
// ends, that should use the terminal.
func (s Shell) runsInline(argv []string) bool {
	if s.sources.Inline == nil {
		return false
	}
	if _, ok := commands.ParseRun(argv); ok {
		return true
	}
	return len(argv) > 0 && cliVerbs[argv[0]]
}

// startInline hands the terminal to the command.
func (s *Shell) startInline(argv []string) tea.Cmd {
	s.lastArg = argv
	if s.history = append(s.history, commands.Line(argv)); len(s.history) > maxHistory {
		s.history = s.history[len(s.history)-maxHistory:]
	}
	c := s.sources.Inline(argv)
	return tea.Exec(c, func(err error) tea.Msg { return inlineDoneMsg{cmd: c, err: err} })
}

// inlineDone records how the run ended and shows the Run tab with the list.
func (s *Shell) inlineDone(msg inlineDoneMsg) tea.Cmd {
	r := msg.cmd.Result
	run := inlineRun{
		Argv: msg.cmd.Argv, Code: r.Exit.Code, Cancelled: r.Exit.Cancelled,
		Failed: r.Exit.Err != nil || msg.err != nil, Took: r.Took,
	}
	_, isRun := commands.ParseRun(run.Argv)
	if isRun {
		s.runs = append(s.runs, run)
		if len(s.runs) > maxHistory {
			s.runs = s.runs[len(s.runs)-maxHistory:]
		}
		s.last = lastOf(run.Argv, run.OK(), run.Cancelled, shortDuration(run.Took))
		if contains(s.tabs, commands.TabRun) {
			s.tab = commands.TabRun
		}
	}
	switch {
	case run.Cancelled:
		s.notify("Stopped: "+commands.Line(run.Argv), toast.Warning)
	case run.OK():
		s.notify("Done in "+shortDuration(run.Took), toast.Success)
	default:
		s.notify(fmt.Sprintf("Failed (exit %d)", run.Code), toast.Error)
	}
	if isRun {
		return s.loadProjects()
	}
	return s.refreshAfter(run.Argv)
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
	return runner.InlineSummary(last.Argv, runner.InlineResult{
		Exit: runner.Exit{Code: last.Code, Cancelled: last.Cancelled},
		Took: last.Took,
	})
}

// inlineRunLines lists the runs, newest first: command, status, duration.
func (s Shell) inlineRunLines(w, h int) []string {
	var lines []string
	for i := len(s.runs) - 1; i >= 0 && len(lines) < h; i-- {
		r := s.runs[i]
		mark, state := s.th.Success("✔"), "passed"
		switch {
		case r.Cancelled:
			mark, state = s.th.Warning("■"), "stopped"
		case !r.OK():
			mark, state = s.th.Failure("✖"), fmt.Sprintf("failed (exit %d)", r.Code)
		}
		line := fmt.Sprintf(" %s %s  %s  %s", mark, strings.TrimPrefix(commands.Line(r.Argv), "smith "), state, s.th.Muted(shortDuration(r.Took)))
		lines = append(lines, line)
	}
	return lines
}
