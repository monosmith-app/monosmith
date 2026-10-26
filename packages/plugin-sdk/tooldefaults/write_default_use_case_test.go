package tooldefaults

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDefaultWritesOnceAndRepairsEdits(t *testing.T) {
	ws := t.TempDir()
	rel, err := WriteDefault(ws, "tool.toml", "a = 1\n")
	if err != nil || rel != ".smith/defaults/tool.toml" {
		t.Fatalf("%q %v", rel, err)
	}
	full := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.WriteFile(full, []byte("edited"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteDefault(ws, "tool.toml", "a = 1\n"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(full)
	if string(got) != "a = 1\n" {
		t.Fatalf("a changed default is restored, got %q", got)
	}
}
