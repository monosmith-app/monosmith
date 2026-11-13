package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestToolchainVersionIsDecidedByToolpin checks the order: smith.toml, go.work,
// then the default.
func TestToolchainVersionIsDecidedByToolpin(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := toolchainVersion(ws, nil); got != "1.25" {
		t.Fatalf("go.work: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\ngo = \"1.26\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := toolchainVersion(ws, nil); got != "1.26" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
