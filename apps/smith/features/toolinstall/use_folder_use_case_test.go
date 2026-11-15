package toolinstall

import (
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

func TestUseFolderRecordsTheWingetFolder(t *testing.T) {
	local := t.TempDir()
	dir := putFile(t, WingetPackagesDir(local), "Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe", "dart-sdk", "bin", "dart.exe")
	file := filepath.Join(t.TempDir(), "tool-folders.json")
	t.Setenv("PATH", t.TempDir())
	got, err := UseFolder(wingetNeed("dart", "Google.DartSDK"), local, file)
	if err != nil || got != dir {
		t.Fatalf("UseFolder = %q, %v; want %q", got, err, dir)
	}
	recorded, _ := toolrun.ReadToolFolders(file)
	if recorded["dart"] != dir {
		t.Fatalf("recorded %v", recorded)
	}
}

func TestUseFolderWithoutAFolderSaysHowToInstall(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tool-folders.json")
	if _, err := UseFolder(wingetNeed("dart", "Google.DartSDK"), t.TempDir(), file); err == nil {
		t.Fatal("no folder must be an error")
	}
	if recorded, _ := toolrun.ReadToolFolders(file); len(recorded) != 0 {
		t.Fatalf("nothing is recorded: %v", recorded)
	}
}
