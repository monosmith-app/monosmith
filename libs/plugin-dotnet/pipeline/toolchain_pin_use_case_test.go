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
	if err := os.WriteFile(filepath.Join(ws, "global.json"), []byte(`{"sdk":{"version":"9.0.100"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "9.0" {
		t.Fatalf("native file: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\ndotnet = \"9.1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "9.1" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
