package tooldefaults

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfigClimbsFromTheProjectToTheWorkspace(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	has := HasFile(".tool.toml")
	if FindConfig(ws, "libs/core", has) {
		t.Fatal("nothing configured yet")
	}
	if err := os.WriteFile(filepath.Join(ws, ".tool.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !FindConfig(ws, "libs/core", has) {
		t.Fatal("the workspace root's file applies to a project below it")
	}
}

func TestFindConfigIgnoresADirectoryOfTheSameName(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, ".tool.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	if FindConfig(ws, "app", HasFile(".tool.toml")) {
		t.Fatal("a directory is not a configuration")
	}
}
