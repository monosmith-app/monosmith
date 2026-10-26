package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeRustc(t *testing.T, out string, err error) {
	t.Helper()
	real := probe
	t.Cleanup(func() { probe = real })
	probe = func(string, []string) (string, error) { return out, err }
}

func TestCheckRustVersionPassesOnTheCIVersion(t *testing.T) {
	fakeRustc(t, "rustc 1.90.1 (abc 2026-01-01)\n", nil)
	f, ok := CheckRustVersion(t.TempDir())
	if !ok || !f.OK || f.Fix != nil {
		t.Fatalf("%v %+v", ok, f)
	}
}

func TestCheckRustVersionFailsOnAnOlderLocalRustAndInstallsTheCIOne(t *testing.T) {
	fakeRustc(t, "rustc 1.76.0 (07dca489a 2024-02-02)\n", nil)
	f, ok := CheckRustVersion(t.TempDir())
	if !ok || f.OK || f.Warning {
		t.Fatalf("%v %+v", ok, f)
	}
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Files) != 1 || f.Fix.Files[0].Path != "rust-toolchain.toml" ||
		!strings.Contains(f.Fix.Files[0].Content, `channel = "1.90"`) || f.Fix.Steps[0].Argv[0] != "rustup" {
		t.Fatalf("%+v", f.Fix)
	}
}

func TestCheckRustVersionWarnsOnANewerLocalRustAndMovesThePin(t *testing.T) {
	fakeRustc(t, "rustc 1.95.2 (abc 2026-09-01)\n", nil)
	f, ok := CheckRustVersion(t.TempDir())
	if !ok || f.OK || !f.Warning {
		t.Fatalf("%v %+v", ok, f)
	}
	if f.Fix == nil || len(f.Fix.Steps) != 0 || !strings.Contains(f.Fix.Files[0].Content, `channel = "1.95"`) {
		t.Fatalf("%+v", f.Fix)
	}
}

func TestCheckRustVersionNeverRewritesAnExistingToolchainFile(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "rust-toolchain.toml"), []byte("[toolchain]\nchannel = \"1.85.0\"\n"), 0o644)
	fakeRustc(t, "rustc 1.76.0 (x)\n", nil)
	f, _ := CheckRustVersion(ws)
	if f.OK || f.Fix == nil || len(f.Fix.Files) != 0 || len(f.Fix.Steps) != 1 {
		t.Fatalf("%+v", f)
	}
	fakeRustc(t, "rustc 1.99.0 (x)\n", nil)
	if f, _ := CheckRustVersion(ws); f.Fix != nil {
		t.Fatalf("a newer local Rust with a pin file has no fix: %+v", f.Fix)
	}
}

func TestCheckRustVersionSaysNothingWithoutRustc(t *testing.T) {
	fakeRustc(t, "", errors.New("not found"))
	if _, ok := CheckRustVersion(t.TempDir()); ok {
		t.Fatal("a missing rustc is the cargo check's finding")
	}
}
