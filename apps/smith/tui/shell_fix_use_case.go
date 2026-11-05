package tui

import (
	"context"
	"errors"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// fixRun is a fix being applied off the UI goroutine. Its steps report on
// ch, one message after the other, and the last message is fixDoneMsg.
type fixRun struct {
	ch     chan tea.Msg
	cancel context.CancelFunc

	mu    sync.Mutex
	lines []string
}

// maxLogLines is how many lines of a step's output the full log keeps; older
// lines drop out.
const maxLogLines = 2000

// Note keeps a line of detail (a file being written, a line a command
// printed). It is not a message: a command that prints fast must not crowd out
// the step messages, and the next frame reads the latest lines.
func (r *fixRun) Note(line string) {
	r.mu.Lock()
	r.lines = append(r.lines, line)
	if len(r.lines) > maxLogLines+maxLogLines/4 {
		r.lines = append([]string(nil), r.lines[len(r.lines)-maxLogLines:]...)
	}
	r.mu.Unlock()
}

// snapshot is the lines kept so far. Note only appends or replaces the slice,
// so the returned lines never change under the reader.
func (r *fixRun) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lines[:len(r.lines):len(r.lines)]
}

// synced is the wizard with the running step's latest lines, and its whole log
// while the full log view is open.
func (s Shell) synced(w fixwizard.Wizard) fixwizard.Wizard {
	if s.fixing == nil {
		return w
	}
	lines := s.fixing.snapshot()
	w = w.Noted(lines)
	if w.LogOpen() {
		w = w.Logged(lines)
	}
	return w
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
		// value is what an adopt job returned.
		value any
	}
)

// stepWriter turns the "… name" lines doctorfix.Apply prints as each step
// starts into progress messages.
type stepWriter struct {
	run *fixRun
	n   int
}

// Note passes a line of detail to the run.
func (w *stepWriter) Note(line string) { w.run.Note(line) }

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
	w, out := s.synced(*s.fix).Key(name)
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
	if s.repair != nil && msg.err == nil {
		// smith.toml reads now: start again so the shell opens the workspace.
		s.fix = nil
		s.reopen, s.quit = true, true
		s.reopenNote = "Fixed smith.toml. " + notCommittedHint()
		return nil
	}
	refresh := s.refreshAfter([]string{"doctor", "--fix"})
	if msg.err != nil {
		w := s.fix.Fail(msg.err)
		s.fix = &w
		return refresh
	}
	what := s.fix.Resolved().Description
	s.verifying = s.fix.Fix.Checks
	s.verifyingWhat = what
	s.fix = nil
	s.notify("Ran: "+what+". Checking that it worked…", toast.Success)
	return refresh
}

// verifyFix runs when doctor has checked again after a fix: every fix verifies
// itself. The toast says "Fixed" only when the checks the fix clears now pass.
// Otherwise it says what doctor found, which file and version for a tool, so
// the person never has to guess why the warning is still there.
func (s *Shell) verifyFix(res doctorview.Result) {
	checks := s.verifying
	s.verifying = nil
	if len(checks) == 0 {
		return
	}
	if res.Err != nil {
		s.notify("The fix ran, but doctor could not check again: "+res.Err.Error(), toast.Warning)
		return
	}
	if still := doctor.StillFailing(res.Report, checks); len(still) > 0 {
		s.notify(doctor.StillFailingText(still[0]), toast.Warning)
		return
	}
	s.notify("Fixed: "+s.verifyingWhat, toast.Success)
}
