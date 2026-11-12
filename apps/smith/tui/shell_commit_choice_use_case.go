package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commandtext"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/resultdialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// capCommit is a capture that writes commit_operations to smith.toml.
const capCommit = "commit-choice"

// commitSettingArgv writes the answer "yes" in smith.toml. --commit makes the
// change one commit of its own, so the work tree stays clean for the next one.
var commitSettingArgv = []string{"config", "set", "workspace.commit_operations", "true", "--commit"}

// commitUnasked is true when the person has not said yet how smith treats undo
// in this workspace: smith.toml does not say, and no answer is saved.
func (s Shell) commitUnasked() bool {
	return s.inWorkspace && !s.commitSet && s.commitChoice == ""
}

// asksCommit is true for a change that smith would commit and has no answer yet:
// the question comes before the confirmation.
func (s Shell) asksCommit(argv []string) bool {
	return s.commitUnasked() && dialog.CanCommit(argv) && !slices.Contains(argv, "--commit")
}

// askCommitChoice shows the question. pending is the change that waits for the
// answer, or nil when the question was opened on its own.
func (s *Shell) askCommitChoice(pending []string, confirmed bool) {
	s.commitPending, s.commitPendingOK = slices.Clone(pending), confirmed
	m := resultdialog.NewCommitChoice(pending != nil).Within(s.dialogMargins()).Place(s.w, s.h)
	s.result = &m
}

// openUndoOff tells why there is nothing to undo when commits are off, with a
// button that opens the question.
func (s *Shell) openUndoOff() {
	m := resultdialog.NewUndoOff().Within(s.dialogMargins()).Place(s.w, s.h)
	s.result = &m
}

// afterCommitChoice does what an answer asks: it saves the answer, then lets the
// waiting change go on to its confirmation.
func (s *Shell) afterCommitChoice(choice resultdialog.CommitChoice) tea.Cmd {
	s.result = nil
	switch choice {
	case resultdialog.CommitOpen:
		s.askCommitChoice(nil, false)
		return nil
	case resultdialog.CommitYes:
		if s.sources.Capture == nil {
			s.notify(words.Say("tui.commit.cannot_save"), toast.Warning)
			return s.resumePending()
		}
		return s.capture(capCommit, slices.Clone(commitSettingArgv))
	}
	if s.saveCommit != nil {
		s.saveCommit(string(choice))
	}
	s.commitChoice = string(choice)
	if choice == resultdialog.CommitNo {
		s.say(words.Say("tui.commit.saved_no"))
	} else {
		s.say(words.Say("tui.commit.saved_ask"))
	}
	return s.resumePending()
}

// commitSaved ends the write of commit_operations: on success undo is on; on a
// failure the reason is shown and the waiting change is dropped.
func (s *Shell) commitSaved(msg capturedMsg) tea.Cmd {
	if !msg.res.OK {
		s.commitPending = nil
		s.tell(words.Say("tui.commit.not_saved"), words.Say("tui.capture.smith_said"), commandtext.Failure(msg.res.Text, msg.res.Err))
		return nil
	}
	s.commitSet, s.commitOps = true, true
	s.say(words.Say("tui.commit.saved_yes"))
	return s.resumePending()
}

// resumePending runs the change that waited for the answer, if any.
func (s *Shell) resumePending() tea.Cmd {
	argv, ok := s.commitPending, s.commitPendingOK
	s.commitPending = nil
	if argv == nil {
		return nil
	}
	return s.execute(argv, ok)
}
