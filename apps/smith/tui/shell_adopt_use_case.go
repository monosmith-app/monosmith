package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

// openAdopt starts the adopt wizard on the folder. Reading the folder is its
// first job, so the dialog opens at once and says it is reading.
func (s *Shell) openAdopt() tea.Cmd {
	if s.adopter == nil {
		s.say("Adoption is not available here")
		return nil
	}
	if s.running() {
		s.say("A run is in progress (Ctrl+C stops it)")
		return nil
	}
	m, job := adoptwizard.New(s.adopter)
	m = m.Place(s.w, s.h)
	s.adopt = &m
	return s.runAdoptJob(job)
}

// adoptKey sends a key to the wizard and does what it asks.
func (s *Shell) adoptKey(name string) tea.Cmd {
	m, out := s.adopt.Key(name)
	s.adopt = &m
	return s.afterAdopt(out)
}

// afterAdopt acts on what the wizard asked for.
func (s *Shell) afterAdopt(out adoptwizard.Outcome) tea.Cmd {
	switch {
	case out.Copy != "":
		s.copyText(out.Copy)
	case out.Stop:
		if s.adopting != nil {
			s.adopting.cancel()
			s.say("Stopping the step…")
		}
	case out.Job != nil:
		return s.runAdoptJob(*out.Job)
	case out.Finished != "":
		s.adopt.Dispose()
		s.adopt = nil
		s.reopen = true
		if out.Build {
			s.reopenStart = []string{"run", "build"}
		}
		s.quit = true
	case out.Close:
		s.adopt.Dispose()
		s.adopt = nil
		if out.Note != "" {
			s.notify(out.Note, toast.Warning)
		} else {
			s.say("Cancelled: nothing was changed")
		}
	}
	return nil
}

// runAdoptJob runs one of the wizard's steps off the UI goroutine. The
// "… name" lines it writes become progress; the end is one message with what
// it returned.
func (s *Shell) runAdoptJob(job adoptwizard.Job) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	run := &fixRun{ch: make(chan tea.Msg, 32), cancel: cancel}
	s.adopting = run
	go func() {
		defer cancel()
		var value any
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%s stopped unexpectedly: %v", job.Name, r)
				}
			}()
			value, err = job.Run(ctx, &stepWriter{run: run})
		}()
		run.ch <- fixDoneMsg{run: run, err: err, value: value}
	}()
	return waitFix(run)
}

// adoptProgress shows the step that started and waits for the next news.
func (s *Shell) adoptProgress(msg fixProgressMsg) tea.Cmd {
	if s.adopt == nil {
		return nil
	}
	m := s.adopt.Progress(msg.n, msg.name)
	s.adopt = &m
	return waitFix(msg.run)
}

// adoptDone ends the job and lets the wizard go on.
func (s *Shell) adoptDone(msg fixDoneMsg) tea.Cmd {
	s.adopting = nil
	if s.adopt == nil {
		return nil
	}
	m, out := s.adopt.Done(msg.value, msg.err)
	s.adopt = &m
	return s.afterAdopt(out)
}

// adoptBanner is what the guide leads with in a folder that has files and is
// not a workspace: the first thing to do. A click on it opens the wizard.
func (s Shell) adoptBanner(w int) []string {
	if s.inWorkspace || s.emptyFolder || s.adopter == nil || s.guide.Title() != "" {
		return nil
	}
	lines := pane.Wrap(s.th.Warning("Not a smith workspace yet: ")+s.th.Bold("Enter starts adoption")+s.th.Muted(" (or press "+commands.ShortcutOf("file.adopt")+")"), w, 0)
	lines = append(lines, pane.Wrap(s.th.Muted("It reads the folder first, asks how far to go, and shows every change before anything is written."), w, 0)...)
	return append(lines, "")
}
