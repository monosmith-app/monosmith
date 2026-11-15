package tui

import "testing"

func TestCommandInIsTheFirstBackticked(t *testing.T) {
	for remedy, want := range map[string]string{
		"run `smith ci generate` and commit": "smith ci generate",
		"install Go":                         "",
		"broken `open":                       "",
	} {
		if got := commandIn(remedy); got != want {
			t.Errorf("commandIn(%q) = %q, want %q", remedy, got, want)
		}
	}
}
