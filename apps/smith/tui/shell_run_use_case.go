package tui

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commandtext"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/logview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releasepick"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/resultdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// maxHistory bounds the command log.
const maxHistory = 200

// execute runs a smith command line the way the user would from a shell:
// commands that need a terminal hand over to it, commands that change the
// workspace or publish ask first in a dialog that shows the line, and the
// rest (task runs included) start at once in the Run tab, which captures
// their output and shows each task's state. confirmed is true once the dialog said yes.
func (s *Shell) execute(argv []string, confirmed bool) tea.Cmd {
	if section, ok := settingsSectionFor(argv); ok && s.inWorkspace {
		// What plugin, recipe and config list print is the Settings tab.
		return s.showSettings(section)
	}
	if isToolsInstall(argv) {
		return s.openTools(argv)
	}
	if len(argv) == 1 && argv[0] == "up" && s.canShow(commands.TabDeps) {
		// Bare `smith up` is a picker for a terminal; the Dependencies tab asks
		// the same questions and runs the chosen update in place.
		return s.gotoTab(commands.TabDeps)
	}
	switch catalog.Classify(argv) {
	case catalog.NeedsTerminal:
		s.Argv, s.quit = argv, true
		return nil
	case catalog.Changes:
		if !confirmed {
			if resultdialog.ChoosesVersion(argv) {
				// Which version wins is the person's choice: ask before the plan.
				m := resultdialog.NewWhichVersion(argv).Place(s.w, s.h)
				s.result = &m
				return nil
			}
			if argv[0] == "release" && !s.release.Loaded && s.release.Err == nil && s.sources.Release != nil && !slices.Contains(argv, "--dry-run") {
				// Publishing cannot be undone: read the plan first, then ask.
				s.releaseWaiting = slices.Clone(argv)
				// The dialog opens at once and fills in when the plan is read. The
				// load is the tab's own, so Esc ignores its result and lets it end.
				title, why := commands.Describe(argv)
				s.planTicket++
				s.openWaiting(title, why, catalog.WithConfirmation(argv), words.Say("tui.run.reading_the_commits_to_work_out"), nil)
				load := s.sources.Release
				return func() tea.Msg { return load() }
			}
			if argv[0] == "release" && s.offersAdjustments(argv) {
				// Which packages go out, and with which bump, is the person's choice: ask before the confirmation.
				m := releasepick.New(s.release.Plan).Within(s.dialogMargins()).Place(s.w, s.h)
				s.relpick = &m
				return nil
			}
			if flag, ok := catalog.PlanFlag(argv); ok && s.sources.Capture != nil {
				// The dialog shows what the command will do, not a sentence about it.
				title, why := commands.Describe(argv)
				return s.captureInDialog(capPlan, title, why, catalog.WithConfirmation(argv), append(slices.Clone(argv), flag), words.Say("tui.run.working_out_what_will_change"))
			}
			s.askConfirm(argv)
			return nil
		}
		argv = catalog.WithConfirmation(argv)
	}
	if !s.inWorkspace {
		if s.sources.Start != nil && runsOutside(argv) {
			// There is no Run tab before a workspace: a dialog shows the output.
			return s.startOutside(argv)
		}
		// Anything else that starts here (smith adopt) runs in the terminal
		// after the shell quits.
		s.Argv, s.quit = argv, true
		return nil
	}
	if s.running() {
		s.say(words.Say("tui.run.a_run_is_in_progress_ctrl"))
		return nil
	}
	return s.start(argv)
}

// askConfirm shows the confirmation of a command that changes the workspace or
// publishes: the sentence, what it touches and the line that will run.
func (s *Shell) askConfirm(argv []string) {
	title, why := commands.Describe(argv)
	c := dialog.NewConfirm(title, why, catalog.WithConfirmation(argv)).WithDetails(s.confirmDetails(argv)).Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
}

// offersAdjustments is true for a plain release whose plan is loaded and has a
// package: the question about holds and bumps comes before the confirmation. A
// line that already has --hold, --bump or --dry-run is the person's own.
func (s Shell) offersAdjustments(argv []string) bool {
	if !s.release.Loaded || len(s.release.Plan.Items) == 0 {
		return false
	}
	return !slices.ContainsFunc(argv, func(a string) bool {
		return a == "--dry-run" || a == "--hold" || a == "--bump" || strings.HasPrefix(a, "--hold=") || strings.HasPrefix(a, "--bump=")
	})
}

// start begins the process; the first output arrives as messages.
func (s *Shell) start(argv []string) tea.Cmd {
	s.run, s.runHome, s.runLines, s.resultKept = runview.NewRun(argv, nil, s.now()), false, nil, nil
	s.formOrigin = ""
	s.remember(s.run)
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
			return startedMsg{argv: argv, err: errors.New(words.Say("tui.run.cannot_start"))}
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
		s.run.Feed(words.Say("tui.run.could_not_start", "err", msg.err.Error()), s.now())
		s.run.End(1, false, s.now())
		s.notify(words.Say("tui.run.could_not_start_2", "err", msg.err.Error()), toast.Error)
		s.showOutside()
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
	s.keepLines(msg.lines)
	if msg.exit == nil {
		s.showOutside()
		return waitCmd(msg.proc)
	}
	exit := *msg.exit
	s.run.End(exit.Code, exit.Cancelled, now)
	s.proc, s.stopping = nil, false
	s.showOutside()
	if s.runSel == "" {
		s.runSel = s.firstFailed()
	}
	s.finished()
	s.afterRun()
	if name, ok := addedProject(s.run.Command); ok && s.run.OK() {
		s.justAdded = name
	}
	s.last = lastOf(s.run.Command, s.run.OK(), s.run.Cancelled, shortDuration(s.run.RunTime(s.now())))
	cmds := []tea.Cmd{s.refreshAfter(s.run.Command)}
	if _, ok := commandtext.ReportTarget(s.run.Command); ok && s.run.OK() {
		cmds = append(cmds, s.reportWritten(s.run.Command))
	}
	return tea.Batch(cmds...)
}

// refreshAfter reads again what a command that is not a task run may have
// changed: the projects, the checks (the header keeps showing the old ones
// until the new arrive), and, when their views are opened, the plan and the
// dependencies.
func (s *Shell) refreshAfter(argv []string) tea.Cmd {
	if _, isRun := commands.ParseRun(argv); isRun {
		return nil
	}
	if isPreview(argv) {
		// A preview writes nothing: the plan on the Release tab stays, so the
		// release can start from it.
		return nil
	}
	s.release, s.deps = releaseview.Model{}, depsview.Model{}
	cmds := []tea.Cmd{s.loadProjects()}
	if len(argv) > 0 && argv[0] == "plugin" {
		// A plugin that came or went changes the choices a project can be built from.
		s.choices, s.offers = choicesNotAsked, nil
		var reload tea.Cmd
		s.settings, reload = s.settings.Reload()
		cmds = append(cmds, reload)
	}
	if load := s.sources.Doctor; load != nil {
		cmds = append(cmds, func() tea.Msg { return load() })
	}
	return tea.Batch(cmds...)
}

// finished tells the user how the run ended.
func (s *Shell) finished() {
	switch {
	case s.run.Cancelled:
		s.notify(words.Say("tui.run.cancelled", "commandLine", s.run.CommandLine()), toast.Warning)
	case s.run.OK() && undoable(s.run.Command, s.commitOps) && !commandtext.SaidNotCommitted(s.run.Lines):
		s.notify(s.summary()+". "+undoHint(), toast.Success)
	case s.run.OK() && changesFiles(s.run.Command):
		s.notify(s.summary()+". "+notCommittedHint(), toast.Success)
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
		if r.Exit == 0 {
			return words.Say("tui.run.done_in", "took", took)
		}
		return words.Say("tui.run.failed_code_in", "exit", r.Exit, "took", took)
	}
	var parts []string
	for _, p := range []struct {
		n    int
		word string
	}{{c.Passed, "passed"}, {c.Cached, "cached"}, {c.Failed, "failed"}, {c.Skipped, "skipped"}} {
		if p.n > 0 {
			parts = append(parts, words.Say("tui.run.text", "n", p.n, "word", p.word))
		}
	}
	return words.Say("tui.run.in", "list", strings.Join(parts, " · "), "took", took)
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
	s.say(words.Say("tui.run.stopping_ctrl_c_again_quits"))
}

// selectedProject is the project of the list's selection, or "".
func (s Shell) selectedProject() string { return s.list.SelectedName() }

func fmtCount(done, total int) string { return fmt.Sprintf("%d/%d", done, total) }
