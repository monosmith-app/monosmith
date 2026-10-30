package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

// A Node project's src-tauri is part of that project, not a crate of its own.
func TestDiscoverLeavesOutATauriShell(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/studio/package.json", "{}")
	write("apps/studio/src-tauri/Cargo.toml", "[package]\nname = \"studio\"\nversion = \"0.1.0\"\n")
	write("apps/studio/src-tauri/src/main.rs", "fn main() {}\n")
	write("libs/src-tauri/Cargo.toml", "[package]\nname = \"plain\"\nversion = \"0.1.0\"\n")
	res, err := Discover(ws, []string{"apps/studio/src-tauri/Cargo.toml", "libs/src-tauri/Cargo.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 || res.Projects[0].Root != "libs/src-tauri" {
		t.Fatalf("only the src-tauri without a Node parent is a crate: %+v", res.Projects)
	}
}
