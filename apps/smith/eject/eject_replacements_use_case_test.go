package eject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

func TestReplacementsShowWhatForceWouldLose(t *testing.T) {
	root := t.TempDir()
	tool := tooldefaults.Tool{Name: "demo", File: "demo.cfg", Target: "demo.cfg"}
	def := defaultPath(root, tool)
	if err := os.MkdirAll(filepath.Dir(def), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(def, []byte("a = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Replacements(root, tool, nil); len(got) != 0 {
		t.Fatalf("no file there, nothing is replaced: %v", got)
	}
	if err := os.WriteFile(filepath.Join(root, "demo.cfg"), []byte("a = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Replacements(root, tool, nil); len(got) != 0 {
		t.Fatalf("the same content loses nothing: %v", got)
	}
	if err := os.WriteFile(filepath.Join(root, "demo.cfg"), []byte("a = 2\nmine = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Replacements(root, tool, nil)
	if len(got) != 1 || got[0].Path != "demo.cfg" || !strings.Contains(got[0].Diff, "-mine = true") {
		t.Fatalf("the diff names the line that is lost: %+v", got)
	}
}
