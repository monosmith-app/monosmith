package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// maxHistory bounds the command log.
const maxHistory = 200

// execute runs a smith command line the way the user would from a shell:
// commands that need a terminal hand over to it, commands that change the
// workspace or publish ask first in a dialog that shows the line, and the
// rest start at once in the Run tab. confirmed is true once the dialog said yes.
func (s *Shell) execute(argv []string, confirmed bool) tea.Cmd {
	switch catalog.Classify(argv) {
	case catalog.NeedsTerminal:
		s.Argv, s.quit = argv, true
		return nil
	case catalog.Changes:
		if !confirmed {
			title, why := commands.Describe(argv)
			c := dialog.NewConfirm(title, why, catalog.WithConfirmation(argv)).WithDetails(s.confirmDetails(argv)).Place(s.w, s.h)
			s.confirm = &c
			return nil
		}
		argv = catalog.WithConfirmation(argv)
	}
	if !s.inWorkspace {
		// No workspace means no Run tab to show the output (smith new, adopt
		// and convert start here): the terminal runs it after the shell quits.
		s.Argv, s.quit = argv, true
		return nil
	}
	if s.running() {
		s.say("A run is in progress (Ctrl+C stops it)")
		return nil
	}
	if s.runsInline(argv) {
		return s.startInline(argv)
	}
	return s.start(argv)
}

// start begins the process; the first output arrives as messages.
func (s *Shell) start(argv []string) tea.Cmd {
	s.run = runview.NewRun(argv, nil, s.now())
	s.proc, s.stopping, s.runSel, s.runTop, s.log, s.lastArg = nil, false, "", 0, logview.Log{}, argv
	if s.history = append(s.history, s.run.CommandLine()); len(s.history) > maxHistory {
		s.history = s.history[len(s.history)-maxHistory:]
	}
	// The preview is shown on the Release screen it was started from.
	stayOnRelease := s.tab == commands.TabRelease && isPreview(argv)
	if contains(s.tabs, commands.TabRun) && !stayOnRelease {
		s.tab = commands.TabRun
	}
	begin := s.sources.Start
	return func() tea.Msg {
		if begin == nil {
			return startedMsg{argv: argv, err: fmt.Errorf("this smith cannot start commands here")}
		}
		p, err := begin(argv)
		return startedMsg{proc: p, argv: argv, err: err}
	}
}

func (s *Shell) started(msg startedMsg) tea.Cmd {
	if s.run == nil || s.run.Finished {
		return nil
	}
	if msg.err != nil {
		s.run.Feed("could not start: "+msg.err.Error(), s.now())
		s.run.End(1, false, s.now())
		s.notify("Could not start: "+msg.err.Error(), toast.Error)
		return nil
	}
	s.proc = msg.proc
	cmds := []tea.Cmd{waitCmd(msg.proc)}
	if plan := s.sources.Plan; plan != nil {
		proc, argv := msg.proc, msg.argv
		cmds = append(cmds, func() tea.Msg { return plannedMsg{proc: proc, ids: plan(argv)} })
	}
	return tea.Batch(cmds...)
}

// waitCmd blocks, off the Update loop, until the process has output or ends.
func waitCmd(p runner.Handle) tea.Cmd {
	return func() tea.Msg {
		lines, exit := p.Next()
		return outputMsg{proc: p, lines: lines, exit: exit}
	}
}

func (s *Shell) output(msg outputMsg) tea.Cmd {
	if s.run == nil || s.proc != msg.proc {
		return nil
	}
	now := s.now()
	for _, l := range msg.lines {
		s.run.Feed(l, now)
	}
	if msg.exit == nil {
		return waitCmd(msg.proc)
	}
	exit := *msg.exit
	s.run.End(exit.Code, exit.Cancelled, now)
	s.proc, s.stopping = nil, false
	if s.runSel == "" {
		for _, t := range s.run.Ordered() {
			if t.Status == runview.Failed {
				s.runSel = t.ID
				break
			}
		}
	}
	s.finished()
	if _, isRun := commands.ParseRun(s.run.Command); !isRun {
		s.doctor, s.release, s.settings = doctorview.Model{}, releaseview.Model{}, settingsview.Model{}
		s.plugins, s.deps, s.recipes = s.plugins.Reset(), depsview.Model{}, s.recipes.Reset()
		if load := s.sources.Projects; load != nil {
			return func() tea.Msg { return load() }
		}
	}
	return nil
}

// finished tells the user how the run ended.
func (s *Shell) finished() {
	switch {
	case s.run.Cancelled:
		s.notify("Cancelled: "+s.run.CommandLine(), toast.Warning)
	case s.run.OK():
		s.notify(s.summary(), toast.Success)
	default:
		s.notify(s.summary(), toast.Error)
	}
}

// summary is the result of the last run in a few words.
func (s *Shell) summary() string {
	r := s.run
	took := shortDuration(r.RunTime(s.now()))
	c := r.Counts()
	if c.Total == 0 {
		return fmt.Sprintf("exit %d in %s", r.Exit, took)
	}
	var parts []string
	for _, p := range []struct {
		n    int
		word string
	}{{c.Passed, "passed"}, {c.Cached, "cached"}, {c.Failed, "failed"}, {c.Skipped, "skipped"}} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", p.n, p.word))
		}
	}
	return strings.Join(parts, " · ") + " in " + took
}

// stop asks the running command to stop; asking again quits the shell.
func (s *Shell) stop() {
	if s.proc == nil {
		return
	}
	if s.stopping {
		s.quit = true
		return
	}
	s.proc.Stop()
	s.stopping = true
	s.say("Stopping… (Ctrl+C again quits)")
}

// selectedProject is the project of the list's selection, or "".
func (s Shell) selectedProject() string { return s.list.SelectedName() }

func fmtCount(done, total int) string { return fmt.Sprintf("%d/%d", done, total) }
