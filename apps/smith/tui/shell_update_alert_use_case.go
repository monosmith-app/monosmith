package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// updateArgv marks the confirmation of the update: the dialog runs nothing in
// the run panel, since a running smith cannot be replaced. Confirming closes
// the interface and runs `smith update --yes` from the terminal.
var updateArgv = []string{"update"} // words:literal

// updateOffered is true while the status bar says an update is available: a
// newer smith is known and the person did not ask to be left alone about it.
func (s Shell) updateOffered() bool { return s.upd.Available && !s.upd.Muted }

// openUpdate opens the dialog that offers the update: from the first-run alert,
// a click on the status bar, or Enter on the Update section. It shows the command
// that will run, and a box that keeps the reminder out of the status bar.
//
// Returns: no command; the dialog is a model of the shell.
func (s *Shell) openUpdate() tea.Cmd {
	if !s.upd.Runnable {
		n := dialog.NewNotice(
			words.Say("tui.update.title", "latest", s.upd.Latest),
			words.Say("tui.update.unknown", "latest", s.upd.Latest, "url", updatenotice.ReleasesURL), nil,
		).Within(s.dialogMargins()).Place(s.w, s.h)
		s.confirm = &n
		return nil
	}
	c := dialog.NewConfirm(
		words.Say("tui.update.title", "latest", s.upd.Latest),
		words.Say("tui.update.why", "current", s.upd.Current, "latest", s.upd.Latest),
		updateArgv,
	).WithRunLabel(words.Say("tui.update.run")).
		WithBox(words.Say("tui.update.mute"), "d"). // words:literal
		WithLine(s.upd.Command).
		Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
	return nil
}

// isUpdateDialog is true for the confirmation opened by openUpdate.
func isUpdateDialog(c dialog.Confirm) bool {
	return !c.Info && len(c.Argv) == 1 && c.Argv[0] == updateArgv[0]
}

// afterUpdate acts on the outcome of the update dialog: a ticked box keeps the
// reminder out of the status bar; Update now leaves the interface and runs
// smith update; anything else leaves the update one click away.
func (s *Shell) afterUpdate(c dialog.Confirm, out dialog.Outcome) tea.Cmd {
	if c.BoxTicked() {
		s.setUpdateMuted(true)
	}
	if out.Confirmed {
		s.Argv, s.quit = []string{"update", "--yes"}, true // words:literal
		return nil
	}
	if s.upd.Muted {
		s.say(words.Say("tui.update.muted"))
		return nil
	}
	s.say(words.Say("tui.update.later"))
	return nil
}

// setUpdateMuted keeps the choice, in the status bar and in the Update section.
// A failed save only leaves the reminder as it was on the next start.
func (s *Shell) setUpdateMuted(muted bool) {
	if s.updMute != nil {
		_ = s.updMute(muted)
	}
	s.upd.Muted = muted
	s.settings = s.settings.WithUpdate(s.upd)
}

// openFirstUpdateAlert opens the alert of a newer smith the first time it is
// known, once per version: it is remembered before it shows, so a person who
// closes it is not asked again for the same version.
func (s *Shell) openFirstUpdateAlert() {
	s.alertPending = false
	if !s.updateOffered() || s.modalOpen() {
		return
	}
	if s.updAlerted != nil {
		s.updAlerted(s.upd.Latest)
	}
	s.openUpdate()
}
