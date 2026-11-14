package settingsview

import (
	"strings"
	"testing"
)

// updateModel is the tab opened on the Update section of a smith that knows a newer one.
func updateModel(t *testing.T, info UpdateInfo, prefs *UpdatePrefs) Model {
	t.Helper()
	b := sample().backend()
	prefs.Info = info
	b.Update = *prefs
	m := New(b, nil)
	m.section = Update
	m, cmd := m.Open()
	m, _ = settle(m, Result{Cmd: cmd})
	m.section = Update
	return m
}

func TestTheUpdateSectionShowsTheNewerSmithAndWhatWouldRun(t *testing.T) {
	m := updateModel(t, UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Command: "npm i -g monosmith@latest", Available: true, Runnable: true}, &UpdatePrefs{})
	out := text(m, 110, 30)
	for _, want := range []string{"Update", "smith 1.0.0", "update available: 2.0.0", "Don't remind me about updates", "Runs: npm i -g monosmith@latest"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestEnterOnTheUpdateRowOffersTheUpdateWhenOneExistsAndAsksTheRegistryWhenNot(t *testing.T) {
	m := updateModel(t, UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Available: true, Runnable: true}, &UpdatePrefs{})
	if _, res := m.Key(keyMsg("enter")); !res.Update {
		t.Fatal("Enter did not offer the update")
	}

	asked := 0
	m = updateModel(t, UpdateInfo{Current: "1.0.0"}, &UpdatePrefs{Check: func() (UpdateInfo, error) {
		asked++
		return UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Available: true, Runnable: true}, nil
	}})
	m, results := press(m, "enter")
	if asked != 1 {
		t.Fatalf("the registry was asked %d times", asked)
	}
	var checked *UpdateInfo
	for _, r := range results {
		if r.Checked != nil {
			checked = r.Checked
		}
	}
	if checked == nil || !checked.Available || !m.upd.Available {
		t.Errorf("the answer did not reach the tab and the shell: %+v", checked)
	}
}

func TestTheReminderBoxSwitchesBackAndForth(t *testing.T) {
	var saved []bool
	m := updateModel(t, UpdateInfo{Current: "1.0.0", Latest: "2.0.0", Available: true, Runnable: true, Muted: true}, &UpdatePrefs{
		SetMuted: func(v bool) error { saved = append(saved, v); return nil },
	})
	if !strings.Contains(text(m, 110, 30), "[x] Don't remind me about updates") {
		t.Fatal("a muted reminder is not ticked")
	}
	m, results := press(m, "down", "enter")
	if len(saved) != 1 || saved[0] || m.upd.Muted {
		t.Fatalf("the box was not unticked: saved=%v muted=%v", saved, m.upd.Muted)
	}
	if last := results[len(results)-1]; !last.MutedChanged && !results[0].MutedChanged {
		t.Error("the shell was not told")
	}
	m, _ = press(m, "enter")
	if len(saved) != 2 || !saved[1] || !m.upd.Muted {
		t.Errorf("the box was not ticked again: saved=%v", saved)
	}
}
