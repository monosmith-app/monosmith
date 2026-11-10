package tui

import (
	"strings"
	"testing"
)

func TestAUnreadableSmithTomlIsToldWithItsPlaceAndAdoptionIsNotOffered(t *testing.T) {
	sh := NewShell(Config{Problem: "smith.toml:3:9: expected value"})
	got := sh.noWorkspaceHint()
	if !strings.Contains(got, "smith.toml:3:9") || strings.Contains(got, "adopt") || strings.Contains(got, "bring this folder in") {
		t.Errorf("hint = %q", got)
	}
}

func TestNoWorkspaceHintNamesTheOneWayIn(t *testing.T) {
	got := Shell{}.noWorkspaceHint()
	if !strings.Contains(got, "smith adopt") || strings.Contains(got, "convert") {
		t.Errorf("hint = %q", got)
	}
}
