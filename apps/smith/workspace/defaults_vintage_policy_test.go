package workspace

import (
	"strings"
	"testing"
)

func TestDefaultsVintageState(t *testing.T) {
	latest := LatestDefaultsVintage()
	for name, tc := range map[string]struct {
		recorded string
		want     VintageState
	}{
		"none recorded": {"", VintageBehind},
		"latest":        {latest, VintageCurrent},
		"older":         {"2020.01", VintageBehind},
		"newer":         {"2999.12", VintageAhead},
	} {
		if got := DefaultsVintageState(tc.recorded); got != tc.want {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}

func TestDefaultsVintageIsReadAndChecked(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\ndefaults = \"2026.11\"\n"))
	if err != nil || cfg.Workspace.Defaults != "2026.11" {
		t.Fatalf("not read: %q, %v", cfg.Workspace.Defaults, err)
	}
	for _, bad := range []string{"latest", "2026", "2026.13", "26.11"} {
		_, err := Parse([]byte("[workspace]\nname = \"w\"\ndefaults = \"" + bad + "\"\n"))
		if err == nil || !strings.Contains(err.Error(), "not a vintage") {
			t.Errorf("%q was accepted: %v", bad, err)
		}
	}
}
