package apptasks

import (
	"path/filepath"
	"testing"
)

func TestAppFoldersAreTheFoldersOfTheBuildAndOfTheElectronWork(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	f, err := AppFolders(ws, filepath.Join(ws, "apps", "api"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Build != filepath.Join(ws, "dist", "apps", "api") || f.ElectronApp != filepath.Join(ws, "dist", "electron", "apps", "api", "app") || f.Name != "api" {
		t.Errorf("%+v", f)
	}
}

func TestAProjectAtTheRootBuildsIntoDist(t *testing.T) {
	ws := t.TempDir()
	f, err := AppFolders(ws, ws)
	if err != nil {
		t.Fatal(err)
	}
	if f.Build != filepath.Join(ws, "dist") {
		t.Errorf("%+v", f)
	}
}

func TestAFolderOutsideTheWorkspaceIsRefused(t *testing.T) {
	base := t.TempDir()
	if _, err := AppFolders(filepath.Join(base, "ws"), filepath.Join(base, "other")); err == nil {
		t.Fatal("a project outside the workspace has no build folder")
	}
}
