package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func linkLab(t *testing.T, info settingsview.UpdateInfo, alert, link bool) *updateLab {
	t.Helper()
	u := &updateLab{}
	u.sh = NewShell(Config{
		Version: "1.0.0", Workspace: "acme", Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}),
		Update: info, UpdateAlert: alert, LinkOffer: link,
		SetUpdateMuted: func(bool) error { return nil }, MarkUpdateAlerted: func(string) {},
	})
	u.send(t, tea.WindowSizeMsg{Width: 110, Height: 30})
	return u
}

func TestTheFirstRunThatLinkedOffersToOpenTheInterfaceAgain(t *testing.T) {
	u := linkLab(t, settingsview.UpdateInfo{Current: "1.0.0"}, false, true)
	view := pane.Strip(u.sh.View())
	for _, want := range []string{"smith now starts directly", "Restart now", "110 ms faster"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in\n%s", want, view)
		}
	}
	u.send(t, keyMsg("y"))
	if !u.sh.Restart || !u.sh.quit {
		t.Errorf("Restart now did not close the interface for a restart: restart=%v quit=%v", u.sh.Restart, u.sh.quit)
	}
}

func TestLaterKeepsTheInterfaceAndSaysHowToUndo(t *testing.T) {
	u := linkLab(t, settingsview.UpdateInfo{Current: "1.0.0"}, false, true)
	u.send(t, keyMsg("esc"))
	if u.sh.Restart || u.sh.quit || u.sh.confirm != nil {
		t.Fatalf("Later changed more than the dialog: restart=%v quit=%v", u.sh.Restart, u.sh.quit)
	}
	if !strings.Contains(pane.Strip(u.sh.View()), "smith link --status") {
		t.Errorf("the hint to see and undo the link is not shown:\n%s", pane.Strip(u.sh.View()))
	}
}

func TestAnUpdateAlertOnScreenGetsAToastAndNoSecondDialog(t *testing.T) {
	info := settingsview.UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Command: "npm i -g monosmith@latest", Available: true, Runnable: true}
	u := linkLab(t, info, true, true)
	view := pane.Strip(u.sh.View())
	if !strings.Contains(view, "smith 2.0.0 is out") || strings.Contains(view, "Restart now") {
		t.Fatalf("two dialogs, or the wrong one first:\n%s", view)
	}
	if !strings.Contains(view, "quit and run smith again") {
		t.Errorf("no toast about the link:\n%s", view)
	}
}

func TestNoOfferWhenThisRunDidNotLink(t *testing.T) {
	u := linkLab(t, settingsview.UpdateInfo{Current: "1.0.0"}, false, false)
	if u.sh.confirm != nil {
		t.Error("a dialog opened though nothing was linked")
	}
}
