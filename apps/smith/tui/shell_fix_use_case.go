package tui

import (
	"context"
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// fixRun is a fix being applied off the UI goroutine. Its steps report on
// ch, one message after the other, and the last message is fixDoneMsg.
type fixRun struct {
	ch     chan tea.Msg
	cancel context.CancelFunc
}

// Messages of a fix that is being applied.
type (
	// fixProgressMsg says a step started: the n-th, named.
	fixProgressMsg struct {
		run  *fixRun
		n    int
		name string
	}
	// fixDoneMsg ends the run; err is nil when every step succeeded.
	fixDoneMsg struct {
		run *fixRun
		err error
	}
)

// stepWriter turns the "… name" lines doctorfix.Apply prints as each step
// starts into progress messages.
type stepWriter struct {
	run *fixRun
	n   int
}

func (w *stepWriter) Write(p []byte) (int, error) {
	for _, line := range strings.Split(string(p), "\n") {
		if name, ok := strings.CutPrefix(line, "… "); ok {
			w.n++
			select {
			case w.run.ch <- fixProgressMsg{run: w.run, n: w.n, name: name}:
			default: // a frame of progress is not worth stalling the step
			}
		}
	}
	return len(p), nil
}

// fixApplying is true while a fix runs.
func (s Shell) fixApplying() bool { return s.fixing != nil }

// openFix starts the wizard on a planned fix.
func (s *Shell) openFix(p doctorfix.Planned) tea.Cmd {
	if s.running() {
		s.say("A run is in progress (Ctrl+C stops it)")
		return nil
	}
	w := fixwizard.New(p, s.sources.ShowFix).Place(s.w, s.h)
	s.fix = &w
	return nil
}

// fixKey sends a key to the wizard and does what it asks.
func (s *Shell) fixKey(name string) tea.Cmd {
	w, out := s.fix.Key(name)
	s.fix = &w
	return s.afterFix(out)
}

// afterFix acts on what the wizard asked for.
func (s *Shell) afterFix(out fixwizard.Outcome) tea.Cmd {
	switch {
	case out.Done:
		failed := s.fix.Phase() == fixwizard.Failed
		s.fix = nil
		if !failed {
			s.say("Cancelled: nothing was changed")
		}
	case out.Apply:
		return s.applyFix()
	case out.Stop:
		if s.fixing != nil {
			s.fixing.cancel()
			s.say("Stopping the step…")
		}
	case out.Copy != "":
		s.copyText(out.Copy)
	}
	return nil
}

// applyFix runs the wizard's resolved fix in the background and reports its
// steps as messages; the interface stays alive.
func (s *Shell) applyFix() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	run := &fixRun{ch: make(chan tea.Msg, 32), cancel: cancel}
	s.fixing = run
	fix := s.fix.Resolved()
	apply := s.sources.ApplyFix
	go func() {
		defer cancel()
		var err error
		if apply == nil {
			err = errors.New("this smith cannot apply fixes here")
		} else {
			err = apply(ctx, fix, &stepWriter{run: run})
		}
		run.ch <- fixDoneMsg{run: run, err: err}
	}()
	return waitFix(run)
}

// waitFix blocks, off the Update loop, until the run has news.
func waitFix(run *fixRun) tea.Cmd {
	return func() tea.Msg { return <-run.ch }
}

// fixProgress shows the step that started and waits for the next news.
func (s *Shell) fixProgress(msg fixProgressMsg) tea.Cmd {
	if msg.run != s.fixing || s.fix == nil {
		return nil
	}
	w := s.fix.Progress(msg.n, len(s.fix.Resolved().Steps), msg.name)
	s.fix = &w
	return waitFix(msg.run)
}

// fixDone ends the run: success closes the wizard with a toast, a failure
// shows the step and its output. Either way doctor checks again, because
// files may have changed.
func (s *Shell) fixDone(msg fixDoneMsg) tea.Cmd {
	if msg.run != s.fixing || s.fix == nil {
		return nil
	}
	s.fixing = nil
	refresh := s.refreshAfter([]string{"doctor", "--fix"})
	if msg.err != nil {
		w := s.fix.Fail(msg.err)
		s.fix = &w
		return refresh
	}
	what := s.fix.Resolved().Description
	s.fix = nil
	s.notify("Fixed: "+what, toast.Success)
	return refresh
}
