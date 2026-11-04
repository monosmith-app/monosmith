package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// planWait is the plan that a confirmation, opened at once, still waits for.
// ticket tells the plan of this dialog from the late plan of one the person
// cancelled; cancel stops the work (it is nil when the work is a shared load).
type planWait struct {
	ticket int
	since  time.Time
	cancel context.CancelFunc
}

// waitsForTicket is true for the captures that open a confirmation first and
// fill it with the plan when the command ends.
func waitsForTicket(purpose string) bool {
	switch purpose {
	case capPlan, capAdd, capSwap, capUndo:
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
	run := s.sources.Capture
	if run == nil {
		s.notify("This smith cannot read that here: run "+commands.Line(plan)+" in a terminal", toast.Warning)
		return nil
	}
	s.closeWaiting()
	ctx, cancel := context.WithCancel(context.Background())
	s.planTicket++
	ticket := s.planTicket
	s.openWaiting(title, why, shown, doing, cancel)
	return func() tea.Msg {
		return capturedMsg{purpose: purpose, argv: plan, res: run(ctx, plan), ticket: ticket}
	}
}

// openWaiting shows the working dialog and starts its clock.
func (s *Shell) openWaiting(title, why string, argv []string, doing string, cancel context.CancelFunc) {
	c := dialog.NewWaiting(title, why, argv, doing).Within(s.dialogMargins()).Place(s.w, s.h)
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
	return true
}

// closeWaiting ends a wait: it stops the work, drops the working dialog and
// forgets a release that waited for its plan.
func (s *Shell) closeWaiting() {
	if w := s.planWait; w != nil {
		if w.cancel != nil {
			w.cancel()
		}
		s.planWait = nil
	}
	if s.confirm != nil && s.confirm.Working() {
		s.confirm = nil
	}
}

// cancelWaiting is Esc on a working dialog: the work stops and nothing opens
// later, even when its plan was about to arrive.
func (s *Shell) cancelWaiting() {
	s.releaseWaiting = nil
	s.closeWaiting()
}

// waitedForPlan is how long the working dialog has waited.
func (s Shell) waitedForPlan() time.Duration {
	if s.planWait == nil {
		return 0
	}
	return s.now().Sub(s.planWait.since)
}
