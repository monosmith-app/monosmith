package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func runNewHere(t *testing.T, folder string, args ...string) (string, error) {
	t.Helper()
	isolatedGit(t)
	t.Chdir(folder)
	restore := newIsInteractive
	newIsInteractive = func() bool { return false }
	t.Cleanup(func() { newIsInteractive = restore })
	cmd := NewCommand("1.0.0")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(append([]string{"--here"}, args...))
	err := cmd.Execute()
	return out.String(), err
}

func TestHereFillsAnEmptyFolderAndNamesItAfterTheFolder(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "acme")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	// Junk a person did not put there does not make the folder non-empty.
	for _, name := range []string{".DS_Store", "Thumbs.db"} {
		if err := os.WriteFile(filepath.Join(folder, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(folder, ".vscode"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := runNewHere(t, folder); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := workspace.Load(folder); err != nil {
		t.Errorf("smith.toml is written in the folder itself: %v", err)
	}
}

func TestHereRefusesAFolderWithFiles(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "acme")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "notes.txt"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runNewHere(t, folder)
	if err == nil || !strings.Contains(err.Error(), "is not empty") || !strings.Contains(err.Error(), "→ smith adopt") {
		t.Fatalf("want a problem with the remedies, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(folder, "smith.toml")); statErr == nil {
		t.Error("nothing is written into a folder that has files")
	}
}

func TestHereAndDirDisagree(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "acme")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runNewHere(t, folder, "--dir", "elsewhere"); err == nil {
		t.Error("--here with --dir is refused")
	}
}
