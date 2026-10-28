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
