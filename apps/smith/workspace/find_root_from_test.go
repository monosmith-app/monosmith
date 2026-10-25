package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRootFromIgnoresTheWorkspaceRootOfTheRunningTask(t *testing.T) {
	task, copyDir := t.TempDir(), t.TempDir()
	for _, dir := range []string{task, copyDir} {
		if err := os.WriteFile(filepath.Join(dir, FileName), []byte("schema = 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("SMITH_WORKSPACE_ROOT", task)
	if got, err := FindRoot(copyDir); err != nil || got != task {
		t.Fatalf("FindRoot follows the task's workspace: got %q, %v", got, err)
	}
	if got, err := FindRootFrom(copyDir); err != nil || got != copyDir {
		t.Fatalf("FindRootFrom must stay on the directory it was given: got %q, %v", got, err)
	}
}
