package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// A pipeline smith generated for a provider the workspace dropped is deleted
// and reported; one the team wrote (no smith header) is kept.
func TestWriteDeletesWhatItNoLongerGenerates(t *testing.T) {
	root := t.TempDir()
	both := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "azure"}}}
	m := model()
	if _, _, err := Write(root, both, m); err != nil {
		t.Fatal(err)
	}
	githubOnly := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	if drift := Drift(root, githubOnly, m); len(drift) != 1 || !strings.Contains(drift[0], "azure-pipelines.yml is no longer configured") {
		t.Fatalf("drift %v", drift)
	}
	changed, _, err := Write(root, githubOnly, m)
	if err != nil || strings.Join(changed, ",") != AzurePath {
		t.Fatalf("changed %v %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, AzurePath)); !os.IsNotExist(err) {
		t.Fatal("the stale pipeline must be deleted")
	}

	_ = os.WriteFile(filepath.Join(root, AzurePath), []byte("# ours\nsteps: []\n"), 0o644)
	if changed, _, _ := Write(root, githubOnly, m); len(changed) != 0 {
		t.Fatalf("a pipeline the team wrote is kept: %v", changed)
	}
}

// Turning the dev container off deletes the one smith generated.
func TestWriteDeletesTheDevcontainerWhenTurnedOff(t *testing.T) {
	root := t.TempDir()
	on := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}, Devcontainer: true}}
	m := model()
	if _, _, err := Write(root, on, m); err != nil {
		t.Fatal(err)
	}
	off := on
	off.CI.Devcontainer = false
	changed, _, err := Write(root, off, m)
	if err != nil || strings.Join(changed, ",") != DevcontainerPath {
		t.Fatalf("changed %v %v", changed, err)
	}
}
