package cargoworkspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceMembers(t *testing.T) {
	ws := t.TempDir()
	roots := []string{"apps/studio"}
	if got := workspaceMembers(ws, roots); got != nil {
		t.Fatalf("no root Cargo.toml: %v", got)
	}
	for body, want := range map[string]int{
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\n":                                        1,
		"[workspace]\nmembers = [\"apps/*/src-tauri\"]\nexclude = [\"apps/studio/src-tauri\"]\n": 0,
		"[workspace]\nmembers = [\"libs/*\", \"apps/*\"]\n":                                      0,
		"[package]\nname = \"x\"\n":                                                              0,
	} {
		write(t, ws, "Cargo.toml", body)
		if got := workspaceMembers(ws, roots); len(got) != want {
			t.Errorf("%q: %v, want %d", body, got, want)
		}
	}
}

func write(t *testing.T, ws, rel, body string) {
	t.Helper()
	p := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
