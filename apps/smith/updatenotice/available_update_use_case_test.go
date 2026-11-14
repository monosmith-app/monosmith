package updatenotice

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCheckAsksNowAndKeepsTheAnswerForTheInterface(t *testing.T) {
	home := t.TempDir()
	o := opts(home, "1.70.0", t0, "1.75.0", nil)
	got, err := Check(context.Background(), o, time.Second)
	if err != nil || !got.Available || got.Latest != "1.75.0" || got.Command != "npm i -g monosmith@latest" {
		t.Fatalf("Check = %+v, %v", got, err)
	}
	// The interface reads the kept answer without a request.
	o.Fetch = func(context.Context) (string, error) { t.Fatal("Available asked the registry"); return "", nil }
	if again, ok := Available(o); !ok || again.Latest != "1.75.0" {
		t.Errorf("Available = %+v, %v", again, ok)
	}
}

func TestCheckReportsAnUnreachableRegistry(t *testing.T) {
	if _, err := Check(context.Background(), opts(t.TempDir(), "1.70.0", t0, "", errors.New("offline")), time.Second); err == nil {
		t.Fatal("no error")
	}
}

func TestCheckIgnoresTheSettingThatTurnsTheDailyCheckOff(t *testing.T) {
	o := opts(t.TempDir(), "1.70.0", t0, "1.75.0", nil)
	o.NoCheck = true
	if got, err := Check(context.Background(), o, time.Second); err != nil || !got.Available {
		t.Fatalf("Check = %+v, %v", got, err)
	}
	if _, ok := Available(o); ok {
		t.Error("Available spoke while the check is off")
	}
}

func TestAvailableIsQuietWhenCurrentOrOnAnotherChannel(t *testing.T) {
	home := t.TempDir()
	update(home, func(m *memory) { m.Latest, m.Channel = "1.70.0", "latest" })
	if _, ok := Available(opts(home, "1.70.0", t0, "", nil)); ok {
		t.Error("said an update exists for the same version")
	}
	update(home, func(m *memory) { m.Latest, m.Channel = "2.0.0-beta.9", "next" })
	if _, ok := Available(opts(home, "1.70.0", t0, "", nil)); ok {
		t.Error("used an answer read for another channel")
	}
}

func TestMutedKeepsTheUpdateKnownAndIsReversible(t *testing.T) {
	home := t.TempDir()
	update(home, func(m *memory) { m.Latest, m.Channel = "1.75.0", "latest" })
	SetMuted(home, true)
	if u, ok := Available(opts(home, "1.70.0", t0, "", nil)); !ok || !u.Muted {
		t.Fatalf("Available = %+v, %v", u, ok)
	}
	SetMuted(home, false)
	if u, _ := Available(opts(home, "1.70.0", t0, "", nil)); u.Muted {
		t.Error("still muted")
	}
}

func TestUpdateCommandFollowsTheChannelAndRunnableTellsAGuessFromAnAnswer(t *testing.T) {
	npm := "/usr/lib/node_modules/monosmith/bin/smith"
	if got := UpdateCommand(npm, "linux", "next"); got != "npm i -g monosmith@next" {
		t.Errorf("next = %q", got)
	}
	if (Update{Command: UpdateCommand("/opt/smith", "linux", "latest")}).Runnable() {
		t.Error("an unknown install is runnable")
	}
	if !(Update{Command: UpdateCommand(npm, "linux", "latest")}).Runnable() {
		t.Error("an npm install is not runnable")
	}
}

func TestTheAlertIsOfferedOncePerVersion(t *testing.T) {
	home := t.TempDir()
	update(home, func(m *memory) { m.Latest, m.Channel = "1.75.0", "latest" })
	if u, _ := Available(opts(home, "1.70.0", t0, "", nil)); u.Alerted {
		t.Fatal("alerted before the alert")
	}
	MarkAlerted(home, "1.75.0")
	if u, _ := Available(opts(home, "1.70.0", t0, "", nil)); !u.Alerted {
		t.Error("not remembered")
	}
	update(home, func(m *memory) { m.Latest = "1.76.0" })
	if u, _ := Available(opts(home, "1.70.0", t0, "", nil)); u.Alerted {
		t.Error("a newer version was taken for the alerted one")
	}
}
