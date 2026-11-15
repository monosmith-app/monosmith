package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/dialog"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// linkArgv marks the confirmation of the restart: it runs nothing in the run
// panel. Confirming closes the interface, and smith opens it again, now through
// the entry that starts the program directly.
var linkArgv = []string{"restart"} // words:literal

// isLinkDialog is true for the confirmation opened by openLinkOffer.
func isLinkDialog(c dialog.Confirm) bool {
	return !c.Info && len(c.Argv) == 1 && c.Argv[0] == linkArgv[0]
}

// openLinkOffer tells that the first run of this version made smith start
// directly (about 110 ms faster and 40 MB less memory per process) and offers to
// open the interface again, since only a new start uses it. Another dialog on
// screen (the update alert) gets a toast instead.
func (s *Shell) openLinkOffer() {
	s.linkPending = false
	if s.modalOpen() {
		s.notify(words.Say("tui.link.toast"), toast.Info)
		return
	}
	c := dialog.NewConfirm(
		words.Say("tui.link.title"),
		words.Say("tui.link.why"),
		linkArgv,
	).WithRunLabel(words.Say("tui.link.run")).
		WithLine(words.Say("tui.link.line")).
		Within(s.dialogMargins()).Place(s.w, s.h)
	s.confirm = &c
}

// afterLink acts on the answer: Restart now closes the interface and asks smith
// to open it again; Later leaves the interface as it is.
func (s *Shell) afterLink(out dialog.Outcome) tea.Cmd {
	if out.Confirmed {
		s.Restart, s.quit = true, true
		return nil
	}
	s.say(words.Say("tui.link.later"))
	return nil
}
