package tauriapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/conf/src-tauri/tauri.conf.json", "{}")
	write(t, ws, "apps/cargo/src-tauri/Cargo.toml", "[package]\nname=\"a\"\n\n[dependencies]\ntauri = { version = \"2\" }\n")
	write(t, ws, "apps/plain/src-tauri/Cargo.toml", "[package]\nname=\"a\"\n\n[dependencies]\nserde = \"1\"\n")
	write(t, ws, "apps/web/package.json", "{}")
	for root, want := range map[string]bool{"apps/conf": true, "apps/cargo": true, "apps/plain": false, "apps/web": false} {
		if got := Detect(ws, root); got != want {
			t.Errorf("%s: %v, want %v", root, got, want)
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
