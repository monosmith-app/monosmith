package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/statusbar"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// updateLab is a shell that knows a newer smith, with what it kept.
type updateLab struct {
	sh      Shell
	muted   []bool
	alerted []string
}

func newUpdateLab(t *testing.T, info settingsview.UpdateInfo, alert bool) *updateLab {
	t.Helper()
	u := &updateLab{}
	u.sh = NewShell(Config{
		Version: info.Current, Workspace: "acme", Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}),
		Update: info, UpdateAlert: alert,
		SetUpdateMuted:    func(m bool) error { u.muted = append(u.muted, m); return nil },
		MarkUpdateAlerted: func(v string) { u.alerted = append(u.alerted, v) },
	})
	u.send(t, tea.WindowSizeMsg{Width: 110, Height: 30})
	return u
}

func (u *updateLab) send(t *testing.T, msg tea.Msg) {
	t.Helper()
	m, _ := u.sh.Update(msg)
	u.sh = m.(Shell)
}

func newer() settingsview.UpdateInfo {
	return settingsview.UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Command: "npm i -g monosmith@latest", Available: true, Runnable: true}
}

func TestTheFirstAlertOffersTheUpdateOnceAndShowsTheCommand(t *testing.T) {
	u := newUpdateLab(t, newer(), true)
	view := pane.Strip(u.sh.View())
	for _, want := range []string{"smith 2.0.0 is out", "Update now", "Don't remind me about updates  (D)", "npm i -g monosmith@latest"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in\n%s", want, view)
		}
	}
	if len(u.alerted) != 1 || u.alerted[0] != "2.0.0" {
		t.Errorf("alerted = %v, the version must be kept before the alert shows", u.alerted)
	}
}

func TestUpdateNowLeavesTheInterfaceAndRunsSmithUpdate(t *testing.T) {
	u := newUpdateLab(t, newer(), true)
	u.send(t, keyMsg("y"))
	if got := strings.Join(u.sh.Argv, " "); got != "update --yes" || !u.sh.quit {
		t.Errorf("argv = %q, quit = %v", got, u.sh.quit)
	}
}

func TestCancelKeepsTheBarAndTheBoxKeepsTheReminderOutOfIt(t *testing.T) {
	u := newUpdateLab(t, newer(), true)
	u.send(t, keyMsg("esc"))
	if u.sh.confirm != nil || u.sh.quit || len(u.muted) != 0 {
		t.Fatalf("cancel changed more than the dialog: %v %v", u.sh.quit, u.muted)
	}
	if !strings.Contains(pane.Strip(u.sh.View()), "update available") {
		t.Fatal("the bar lost the update after Cancel")
	}

	u = newUpdateLab(t, newer(), true)
	u.send(t, keyMsg("d"))
	u.send(t, keyMsg("esc"))
	if len(u.muted) != 1 || !u.muted[0] || !u.sh.upd.Muted {
		t.Fatalf("muted = %v", u.muted)
	}
	if strings.Contains(pane.Strip(u.sh.View()), "update available") {
		t.Error("the bar still says update available after Don't remind me")
	}
}

func TestAClickOnTheBarMessageOpensTheUpdateAgain(t *testing.T) {
	u := newUpdateLab(t, newer(), false)
	if u.sh.confirm != nil {
		t.Fatal("a dialog opened without the alert")
	}
	first, _, ok := statusbar.UpdateSpan(u.sh.th, u.sh.statusInfo(), u.sh.w)
	if !ok {
		t.Fatal("no clickable span")
	}
	u.send(t, click(first+1, u.sh.h-1))
	u.send(t, letGo(first+1, u.sh.h-1))
	if u.sh.confirm == nil || !strings.Contains(pane.Strip(u.sh.View()), "smith 2.0.0 is out") {
		t.Fatal("the click did not open the update")
	}
}

func TestAMutedUpdateStaysOutOfTheBarAndAnUnknownInstallOnlyTells(t *testing.T) {
	muted := newer()
	muted.Muted = true
	u := newUpdateLab(t, muted, false)
	if strings.Contains(pane.Strip(u.sh.View()), "update available") {
		t.Error("a muted update shows in the bar")
	}
	unknown := newer()
	unknown.Runnable = false
	u = newUpdateLab(t, unknown, false)
	u.sh.openUpdate()
	if u.sh.confirm == nil || !u.sh.confirm.Info {
		t.Fatal("an install smith cannot update must only tell")
	}
}
