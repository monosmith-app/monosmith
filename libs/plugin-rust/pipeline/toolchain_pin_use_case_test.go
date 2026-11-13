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
	if err := os.WriteFile(filepath.Join(ws, "rust-toolchain.toml"), []byte("[toolchain]\nchannel = \"1.80\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "1.80" {
		t.Fatalf("native file: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\nrustc = \"1.85\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Version(ws); got != "1.85" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
