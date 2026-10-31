package cargoworkspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExcludeFromCargoWorkspace(t *testing.T) {
	for body, want := range map[string]string{
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\nexclude = [\"libs/x\"]\n":        `exclude = ["libs/x", "apps/studio/src-tauri"]`,
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\nexclude = [\n  \"libs/x\",\n]\n": "\"libs/x\", \"apps/studio/src-tauri\"\n]",
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\nexclude = []\n":                  `exclude = ["apps/studio/src-tauri"]`,
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\n":                                "exclude = [\"apps/studio/src-tauri\"]",
	} {
		ws := t.TempDir()
		write(t, ws, "Cargo.toml", body)
		if changed, err := ExcludeFromCargoWorkspace(ws, "apps/studio"); err != nil || !changed {
			t.Fatalf("%q: %v %v", body, changed, err)
		}
		got, _ := os.ReadFile(filepath.Join(ws, "Cargo.toml"))
		if !strings.Contains(string(got), want) || len(workspaceMembers(ws, []string{"apps/studio"})) != 0 {
			t.Errorf("%q gave %q, want %q", body, got, want)
		}
		if changed, _ := ExcludeFromCargoWorkspace(ws, "apps/studio"); changed {
			t.Errorf("%q: a second run changes nothing", body)
		}
	}
}
