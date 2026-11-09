package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// openAdopt starts the adopt wizard on the folder. Reading the folder is its
// first job, so the dialog opens at once and says it is reading.
func (s *Shell) openAdopt() tea.Cmd {
	if s.adopter == nil {
		s.say(words.Say("tui.adopt.adoption_is_not_available_here"))
		return nil
	}
	if s.running() {
		s.say(words.Say("tui.adopt.a_run_is_in_progress_ctrl"))
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
			s.say(words.Say("tui.adopt.stopping_the_step"))
		}
	case out.Job != nil:
		return s.runAdoptJob(*out.Job)
	case out.Finished != "":
		s.adopt.Dispose()
		s.adopt = nil
		s.reopen = true
		s.reopenNote = words.Say("tui.adopt.brought_into_monosmith", "notCommittedHint", notCommittedHint())
		if out.Committed {
			s.reopenNote = words.Say("tui.adopt.brought_into_monosmith_2", "undoHint", undoHint())
		}
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
			s.say(words.Say("tui.adopt.cancelled_nothing_was_changed"))
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
					err = errors.New(words.Say("tui.adopt.stopped_unexpectedly", "name", job.Name, "reason", r))
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

// brokenConfigBanner is what the guide leads with when smith.toml exists but
// cannot be read: the whole message, wrapped, and the one way on.
func (s Shell) brokenConfigBanner(w int) []string {
	if s.inWorkspace || s.problem == "" {
		return nil
	}
	lines := pane.Wrap(s.th.Warning(words.Say("tui.adopt.smith_toml_cannot_be_read"))+s.problem, w, 0)
	if s.repair != nil {
		lines = append(lines, pane.Wrap(s.th.Bold(words.Say("tui.adopt.enter_on_fix_smith_toml_shows"))+s.th.Muted(words.Say("tui.adopt.nothing_is_written_until_you_confirm")), w, 0)...)
		return append(lines, "")
	}
	lines = append(lines, pane.Wrap(s.th.Muted(words.Say("tui.adopt.fix_that_line_then_run_smith")), w, 0)...)
	return append(lines, "")
}

// adoptBanner is what the guide leads with in a folder that has files and is
// not a workspace: the first thing to do. A click on it opens the wizard.
func (s Shell) adoptBanner(w int) []string {
	if s.inWorkspace || s.emptyFolder || s.adopter == nil || s.guide.Title() != "" {
		return nil
	}
	lines := pane.Wrap(s.th.Warning(words.Say("tui.adopt.not_a_smith_workspace_yet"))+s.th.Bold(words.Say("tui.adopt.enter_brings_it_in"))+s.th.Muted(words.Say("tui.adopt.or_press", "shortcutOf", commands.ShortcutOf("file.adopt"))), w, 0)
	lines = append(lines, pane.Wrap(s.th.Muted(words.Say("tui.adopt.it_reads_the_folder_first_asks")), w, 0)...)
	return append(lines, "")
}
