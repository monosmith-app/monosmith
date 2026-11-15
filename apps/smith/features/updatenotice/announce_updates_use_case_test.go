package updatenotice

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)

func opts(home, version string, now time.Time, latest string, err error) Options {
	return Options{
		Home: home, Version: version, Exe: "/usr/lib/node_modules/monosmith/bin/smith", GOOS: "linux",
		Now: func() time.Time { return now }, Wait: time.Second,
		Fetch: func(context.Context) (string, error) { return latest, err },
	}
}

func TestFirstRunSaysNothingAndRemembers(t *testing.T) {
	home := t.TempDir()
	if lines := Start(opts(home, "1.70.0", t0, "1.70.0", nil)).Lines(); len(lines) != 0 {
		t.Fatalf("lines = %v", lines)
	}
	if got := load(home).LastVersion; got != "1.70.0" {
		t.Errorf("remembered %q", got)
	}
}

func TestUpgradeIsAnnouncedOnce(t *testing.T) {
	home := t.TempDir()
	Start(opts(home, "1.70.0", t0, "1.70.0", nil)).Lines()
	lines := Start(opts(home, "1.73.0", t0, "1.73.0", nil)).Lines()
	if len(lines) != 1 || !strings.Contains(lines[0], "smith updated 1.70.0 → 1.73.0") || !strings.Contains(lines[0], "releases/tag/smith@1.73.0") {
		t.Fatalf("lines = %v", lines)
	}
	if lines := Start(opts(home, "1.73.0", t0, "1.73.0", nil)).Lines(); len(lines) != 0 {
		t.Errorf("announced twice: %v", lines)
	}
}

func TestNewerOnRegistryIsSaidOncePerDay(t *testing.T) {
	home := t.TempDir()
	lines := Start(opts(home, "1.70.0", t0, "1.75.0", nil)).Lines()
	if len(lines) != 1 || !strings.Contains(lines[0], "smith 1.75.0 is out, you have 1.70.0: update with npm i -g monosmith@latest") {
		t.Fatalf("lines = %v", lines)
	}
	if lines := Start(opts(home, "1.70.0", t0.Add(time.Hour), "1.75.0", nil)).Lines(); len(lines) != 0 {
		t.Errorf("said twice in a day: %v", lines)
	}
	if lines := Start(opts(home, "1.70.0", t0.Add(25*time.Hour), "1.75.0", nil)).Lines(); len(lines) != 1 {
		t.Errorf("not said the next day: %v", lines)
	}
}

func TestOfflineIsSilentAndNotRetriedAllDay(t *testing.T) {
	home := t.TempDir()
	calls := 0
	o := opts(home, "1.70.0", t0, "", errors.New("offline"))
	o.Fetch = func(context.Context) (string, error) { calls++; return "", errors.New("offline") }
	if lines := Start(o).Lines(); len(lines) != 0 {
		t.Fatalf("lines = %v", lines)
	}
	Start(o).Lines()
	if calls != 1 {
		t.Errorf("asked the registry %d times", calls)
	}
}

func TestTurnedOffAsksNothing(t *testing.T) {
	for name, change := range map[string]func(*Options){"ci": func(o *Options) { o.Off = true }, "env": func(o *Options) { o.NoCheck = true }} {
		o := opts(t.TempDir(), "1.70.0", t0, "1.75.0", nil)
		o.Fetch = func(context.Context) (string, error) { t.Errorf("%s: asked the registry", name); return "", nil }
		change(&o)
		if lines := Start(o).Lines(); len(lines) != 0 {
			t.Errorf("%s: lines = %v", name, lines)
		}
	}
}

func TestDevelopmentBuildSaysNothing(t *testing.T) {
	if lines := Start(opts(t.TempDir(), "dev", t0, "1.75.0", nil)).Lines(); len(lines) != 0 {
		t.Errorf("lines = %v", lines)
	}
}

func TestNextChannelHearsAboutNextOnly(t *testing.T) {
	home := t.TempDir()
	var asked string
	o := opts(home, "2.0.0-beta.22", t0, "2.0.0-beta.23", nil)
	o.Fetch = func(context.Context) (string, error) { asked = "called"; return "2.0.0-beta.23", nil }
	lines := Start(o).Lines()
	if asked == "" || len(lines) != 1 || !strings.Contains(lines[0], "smith 2.0.0-beta.23 is out on the next channel, you have 2.0.0-beta.22: update with npm i -g monosmith@next") {
		t.Fatalf("lines = %v", lines)
	}
	if m := load(home); m.Channel != "next" {
		t.Errorf("remembered channel %q", m.Channel)
	}
	// A stable smith must not use the answer read for the next channel.
	if lines := Start(opts(home, "1.70.0", t0.Add(48*time.Hour), "", errors.New("offline"))).Lines(); len(lines) != 0 {
		t.Errorf("a next answer reached a stable smith: %v", lines)
	}
}

func TestNotChosenAnnouncesTheRequestAndAsksNothingForADay(t *testing.T) {
	home := t.TempDir()
	calls := 0
	o := opts(home, "1.70.0", t0, "1.75.0", nil)
	o.Announce = true
	o.Fetch = func(context.Context) (string, error) { calls++; return "1.75.0", nil }
	lines := Start(o).Lines()
	if len(lines) != 1 || !strings.Contains(lines[0], "one request to registry.npmjs.org") || !strings.Contains(lines[0], `update_check = "on" or "off"`) || calls != 0 {
		t.Fatalf("lines = %v, calls = %d", lines, calls)
	}
	o.Now = func() time.Time { return t0.Add(time.Hour) }
	if lines := Start(o).Lines(); len(lines) != 0 || calls != 0 {
		t.Fatalf("announced or asked again within a day: %v, calls = %d", lines, calls)
	}
	o.Now = func() time.Time { return t0.Add(25 * time.Hour) }
	if lines := Start(o).Lines(); len(lines) != 1 || !strings.Contains(lines[0], "smith 1.75.0 is out") || calls != 1 {
		t.Fatalf("not asked after a day: %v, calls = %d", lines, calls)
	}
}

func TestAlwaysSaysANewerVersionEveryTime(t *testing.T) {
	home := t.TempDir()
	Start(opts(home, "1.70.0", t0, "1.75.0", nil)).Lines()
	o := opts(home, "1.70.0", t0.Add(time.Hour), "1.75.0", nil)
	o.Always = true
	if lines := Start(o).Lines(); len(lines) != 1 {
		t.Errorf("--version or doctor must repeat the line: %v", lines)
	}
}

func TestSettingFromPrefersTheUser(t *testing.T) {
	for _, c := range []struct{ env, file, want string }{
		{"", "", ""}, {"on", "", "on"}, {"", `"off"`, "off"}, {"off", `"on"`, "off"}, {"maybe", `"on"`, "on"}, {"maybe", "", ""},
	} {
		if got := SettingFrom(c.env, c.file); got != c.want {
			t.Errorf("SettingFrom(%q, %q) = %q, want %q", c.env, c.file, got, c.want)
		}
	}
}
