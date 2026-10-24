package sources

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAFolderWithNothingButGitIsEmpty(t *testing.T) {
	dir := t.TempDir()
	if !FolderIsEmpty(dir) {
		t.Fatal("a new folder is empty")
	}
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !FolderIsEmpty(dir) {
		t.Fatal(".git alone holds nothing to adopt")
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if FolderIsEmpty(dir) {
		t.Fatal("a manifest is something to adopt")
	}
}
