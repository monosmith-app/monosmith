package tui

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/resultdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// maxRunLines bounds the output kept for the result dialog.
const maxRunLines = 600

// keepLines remembers the output of the run that is going, so its end can be
// read: a failed add says in it which tool is missing.
func (s *Shell) keepLines(lines []string) {
	s.runLines = append(s.runLines, lines...)
	if over := len(s.runLines) - maxRunLines; over > 0 {
		s.runLines = slices.Clone(s.runLines[over:])
	}
}

// notCommittedMark is how a finished change says its files are not committed.
func notCommittedMark() string { return words.Say("tui.result.not_committed") }

// afterRun opens the result dialog for a run that just ended: a failed add that
// left its project on disk because a tool is missing, or a change that was not
// committed. Anything else keeps the toast and the run panel as they are.
func (s *Shell) afterRun() {
	if s.run == nil || s.run.Cancelled || s.result != nil || s.tools != nil || s.confirm != nil {
		return
	}
	argv := s.run.Command
	if !s.run.OK() {
		if tool, ok := resultdialog.ParseMissingTool(argv, s.runLines); ok {
			m := resultdialog.NewMissingTool(tool).Within(s.dialogMargins()).Place(s.w, s.h)
			s.result, s.resultKept = &m, &m
		}
		return
	}
	if slices.ContainsFunc(s.runLines, func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), notCommittedMark()) }) {
		m := resultdialog.NewAfterChange(argv, s.runLines).Within(s.dialogMargins()).Place(s.w, s.h)
		s.result = &m
	}
}

// resultKey sends a key to the result dialog and does what it asks. A commit,
// an install and a new check go through the usual confirmation or dialog, so
// nothing runs because a key was pressed in this one.
func (s *Shell) resultKey(name string) tea.Cmd {
	m, out := s.result.Key(name)
	s.result = &m
	return s.afterResult(out)
}

// resultClick sends a click to the result dialog: one press picks a choice, a
// double press selects it like Enter.
func (s *Shell) resultClick(x, y int) tea.Cmd {
	m, out := s.result.Click(s.th, x, y)
	s.result = &m
	return s.afterResult(out)
}

// afterResult does what the result dialog asked for.
func (s *Shell) afterResult(out resultdialog.Outcome) tea.Cmd {
	if cmd, handled := s.outsideOutcome(out); handled {
		return cmd
	}
	switch {
	case out.CommitChoice != "":
		return s.afterCommitChoice(out.CommitChoice)
	case out.Commit != nil:
		s.result, s.resultKept = nil, nil
		return s.execute(out.Commit, false)
	case out.Install != nil:
		s.result = nil
		return s.execute(out.Install, false)
	case out.Check != nil:
		s.result, s.resultKept = nil, nil
		// A cache clear chosen in the dialog goes on to its confirmation,
		// without asking which cache again.
		s.cacheTierChosen = true
		defer func() { s.cacheTierChosen = false }()
		return s.execute(out.Check, false)
	case out.Open != "":
		s.copyText(out.Open)
		if open := s.sources.Open; open != nil {
			link := out.Open
			return func() tea.Msg { return openedMsg{path: link, link: true, err: open(link)} }
		}
	case out.Done:
		s.result, s.resultKept, s.commitPending = nil, nil, nil
		if out.Note != "" {
			s.notify(out.Note, toast.Info)
		}
	}
	return nil
}

// reopenResult shows the failure dialog again, on the choice after the install
// it sent the person to; it does nothing when no failure dialog was kept.
func (s *Shell) reopenResult() {
	if s.resultKept == nil || s.result != nil {
		return
	}
	m := s.resultKept.Within(s.dialogMargins()).Place(s.w, s.h)
	s.result = &m
}
