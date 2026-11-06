package tui

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/resultdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// runsOutside reports whether a command line runs inside the shell before a
// workspace exists: smith new and smith convert, which make the workspace.
// smith adopt has its own wizard.
func runsOutside(argv []string) bool {
	return len(argv) > 0 && (argv[0] == "new" || argv[0] == "convert")
}

// startOutside starts smith new or smith convert as a child process and opens
// the dialog that shows its output, because no Run tab exists yet.
func (s *Shell) startOutside(argv []string) tea.Cmd {
	if s.running() {
		s.say(words.Say("tui.outside.a_run_is_in_progress_ctrl"))
		return nil
	}
	cmd := s.start(argv)
	s.showOutside()
	return cmd
}

// showOutside draws the output of the run that goes (or went) in the dialog;
// it does nothing inside a workspace, where the Run tab shows it.
func (s *Shell) showOutside() {
	if s.inWorkspace || s.run == nil || !runsOutside(s.run.Command) {
		return
	}
	state := resultdialog.Running
	switch {
	case s.run.Finished && s.run.OK():
		state = resultdialog.Succeeded
	case s.run.Finished:
		state = resultdialog.Failed
	}
	m := resultdialog.NewRunOutput(s.run.Command, s.runLines, state).Place(s.w, s.h)
	s.result = &m
}

// outsideElapsed is how long the command in the output dialog has run.
func (s *Shell) outsideElapsed() time.Duration {
	if s.run == nil || s.run.Finished {
		return 0
	}
	return s.now().Sub(s.run.Started)
}

// outsideOutcome does what the output dialog asks for: stop the command, open
// the workspace it made, or go back to the form with the answers kept. handled
// is false for any other dialog.
func (s *Shell) outsideOutcome(out resultdialog.Outcome) (cmd tea.Cmd, handled bool) {
	if s.inWorkspace || s.run == nil || !runsOutside(s.run.Command) {
		return nil, false
	}
	switch {
	case out.Stop, out.Done && s.running():
		// Esc on a running command stops it: the box stays until it ends.
		s.stop()
		return nil, true
	case out.Enter != "":
		s.result, s.resultKept = nil, nil
		s.reopen, s.reopenDir, s.quit = true, out.Enter, true
		s.reopenNote = words.Say("tui.outside.workspace_ready", "outsideHint", s.outsideHint())
		return nil, true
	case out.Done:
		failed := s.run.Finished && !s.run.OK()
		s.result, s.resultKept = nil, nil
		if failed && s.Form != "" {
			if g, ok := s.guide.Resume(s.Form, s.Answers, s.failureLine()); ok {
				s.guide = g
			}
		}
		return nil, true
	}
	return nil, false
}

// outsideHint says what happens to the files the finished command wrote.
func (s *Shell) outsideHint() string {
	switch {
	case s.run.Command[0] == "new":
		return words.Say("tui.outside.the_first_commit_is_made")
	case slices.Contains(s.run.Command, "--commit"):
		return undoHint()
	}
	return notCommittedHint()
}

// failureLine is the last line the failed command printed, without the
// "smith:" prefix: the reason the form shows under its first question.
func (s *Shell) failureLine() string {
	for i := len(s.runLines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(s.runLines[i]); l != "" {
			return strings.TrimPrefix(l, "smith: ") // words:literal
		}
	}
	return words.Say("tui.outside.the_command_did_not_finish")
}
