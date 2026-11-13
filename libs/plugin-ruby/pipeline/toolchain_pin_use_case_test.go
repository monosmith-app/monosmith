package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestVersionIsDecidedByToolpin checks the order: smith.toml, the native pin
// file, then the default.
func TestVersionIsDecidedByToolpin(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, ".ruby-version"), []byte("3.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "3.1" {
		t.Fatalf("native file: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\nruby = \"3.2\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "3.2" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
