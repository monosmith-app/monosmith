package tauriadd

import (
	"slices"
	"testing"
)

func TestIconCommandRunsTheProjectsOwnCLIOnTheLogo(t *testing.T) {
	want := []string{"npx", "--no-install", "tauri", "icon", "/w/logo.png"}
	if got := iconCommand("/w/logo.png"); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
