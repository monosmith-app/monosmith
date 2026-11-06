package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// planWait is the plan that a confirmation, opened at once, still waits for.
// ticket tells the plan of this dialog from the late plan of one the person
// cancelled; cancel stops the work (it is nil when the work is a shared load).
type planWait struct {
	ticket int
	since  time.Time
	cancel context.CancelFunc
	// purpose is what the wait reads, so Esc can forget what the request set in
	// the shell (a "reading" flag); note is what the status bar says after Esc.
	// The note is empty for the confirmation of a plan, which has its own words.
	purpose string
	note    string
}

// waitsForTicket is true for the captures that open a confirmation first and
// fill it with the plan when the command ends.
func waitsForTicket(purpose string) bool {
	switch purpose {
	case capPlan, capAdd, capSwap, capUndo, capChoices, capRecipeLevels, capSearch, capTools, capCompose:
		return true
	}
	return false
}

// captureInDialog opens the confirmation of a command at once, in its working
// state (a sentence, a spinner and the command, with Run not yet available),
// and reads the plan of the command in the background. The reader of the
// purpose fills the dialog when the plan arrives; Esc stops the command.
//
// Parameters: purpose names the reader; title and why are the words of the
// dialog; shown is the command line the dialog will run; plan is the command
// line that reads the plan; doing says what smith does now.
// Returns: the command that reads the plan, or nil when this smith cannot.
func (s *Shell) captureInDialog(purpose, title, why string, shown, plan []string, doing string) tea.Cmd {
	return s.captureWorking(purpose, title, why, shown, plan, doing, "")
}

// captureWorking is captureInDialog for a read that opens a screen and not a
// confirmation: the working dialog says what is read, and Esc stops the read.
//
// Parameters: as captureInDialog; note is what the status bar says after Esc.
// Returns: the command that reads, or nil when this smith cannot.
func (s *Shell) captureWorking(purpose, title, why string, shown, plan []string, doing, note string) tea.Cmd {
	run := s.sources.Capture
	if run == nil {
		s.notify(words.Say("tui.waiting.this_smith_cannot_read_that_here", "line", commands.Line(plan)), toast.Warning)
		return nil
	}
	s.closeWaiting()
	ctx, cancel := context.WithCancel(context.Background())
	s.planTicket++
	ticket := s.planTicket
	working := dialog.NewWaiting(title, why, shown, doing)
	if note != "" {
		working = working.Reading()
	}
	s.openWaitingAs(working, cancel)
	s.planWait.purpose, s.planWait.note = purpose, note
	return func() tea.Msg {
		return capturedMsg{purpose: purpose, argv: plan, res: run(ctx, plan), ticket: ticket}
	}
}

// openWaiting shows the working dialog and starts its clock.
func (s *Shell) openWaiting(title, why string, argv []string, doing string, cancel context.CancelFunc) {
	s.openWaitingAs(dialog.NewWaiting(title, why, argv, doing), cancel)
}

// openWaitingAs shows a working dialog that was built, and starts its clock.
func (s *Shell) openWaitingAs(c dialog.Confirm, cancel context.CancelFunc) {
	c = c.Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	s.planWait = &planWait{ticket: s.planTicket, since: s.now(), cancel: cancel}
}

// keepsPlan is true when the plan of a finished capture is still wanted: the
// dialog that waited for it is open and was not cancelled. It ends the wait.
func (s *Shell) keepsPlan(msg capturedMsg) bool {
	w := s.planWait
	if w == nil || w.ticket != msg.ticket {
		return false
	}
	w.cancel()
	s.planWait = nil
	if s.confirm != nil && s.confirm.Working() {
		// The reader of the answer opens what comes next.
		s.confirm = nil
	}
	return true
}

// closeWaiting ends a wait: it stops the work, drops the working dialog and
// forgets a release that waited for its plan.
func (s *Shell) closeWaiting() {
	if w := s.planWait; w != nil {
		if w.cancel != nil {
			w.cancel()
		}
		s.forget(w.purpose)
		s.planWait = nil
	}
	if s.confirm != nil && s.confirm.Working() {
		s.confirm = nil
	}
}

// forget puts the shell back as it was before a read that was cancelled began:
// the next press asks again.
func (s *Shell) forget(purpose string) {
	switch purpose {
	case capChoices:
		s.choices, s.formPending = choicesNotAsked, ""
	case capTools:
		s.toolsArgv = nil
	}
}

// cancelWaiting is Esc on a working dialog: the work stops and nothing opens
// later, even when its plan was about to arrive.
func (s *Shell) cancelWaiting() {
	s.releaseWaiting = nil
	s.closeWaiting()
}

// waitNote is what the status bar says when Esc ends the wait now: "" for the
// confirmation of a plan, which says what was not run.
func (s Shell) waitNote() string {
	if s.planWait == nil {
		return ""
	}
	return s.planWait.note
}

// waitedForPlan is how long the working dialog has waited.
func (s Shell) waitedForPlan() time.Duration {
	if s.planWait == nil {
		return 0
	}
	return s.now().Sub(s.planWait.since)
}
