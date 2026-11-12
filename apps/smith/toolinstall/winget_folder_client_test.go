package toolinstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// putFile writes an empty file and its folders under a fake LOCALAPPDATA.
func putFile(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(parts...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(path)
}

func wingetNeed(program, pkg string) Need {
	return Need{ToolNeed: protocol.ToolNeed{Name: program, Program: program, Install: []protocol.ToolInstall{{Manager: "brew", Package: program}, {Manager: "winget", Package: pkg}}}}
}

func TestFindWingetFolder(t *testing.T) {
	local := t.TempDir()
	packages := WingetPackagesDir(local)
	dart := putFile(t, packages, "Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe", "dart-sdk", "bin", "dart.exe")
	putFile(t, packages, "Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe", "dart-sdk", "bin", "utils", "dart.exe")
	php := putFile(t, packages, "PHP.PHP.8.4_Microsoft.Winget.Source_8wekyb3d8bbwe", "php.exe")
	putFile(t, packages, "Deep.Tool_Microsoft.Winget.Source_8wekyb3d8bbwe", "a", "b", "c", "d", "deep.exe")
	putFile(t, packages, "Google.DartSDKBeta_Microsoft.Winget.Source_8wekyb3d8bbwe", "bin", "dart.exe")

	cases := []struct {
		name string
		need Need
		want string
	}{
		{"the Dart SDK zip, nearest bin folder", wingetNeed("dart", "Google.DartSDK"), dart},
		{"a version placeholder matches the installed version", wingetNeed("php", "PHP.PHP.{major.minor}"), php},
		{"too deep is not found", wingetNeed("deep", "Deep.Tool"), ""},
		{"no package folder", wingetNeed("ninja", "Ninja-build.Ninja"), ""},
		{"no winget way", Need{ToolNeed: protocol.ToolNeed{Program: "dart", Install: []protocol.ToolInstall{{Manager: "brew", Package: "dart"}}}}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FindWingetFolder(local, c.need); got != c.want {
				t.Fatalf("FindWingetFolder = %q, want %q", got, c.want)
			}
		})
	}
	if FindWingetFolder("", wingetNeed("dart", "Google.DartSDK")) != "" {
		t.Fatal("no LOCALAPPDATA, no folder")
	}
}

func TestFindWingetFolderTakesAnAlternative(t *testing.T) {
	local := t.TempDir()
	dir := putFile(t, WingetPackagesDir(local), "Python.Python.3.12_src", "python3.exe")
	need := wingetNeed("python", "Python.Python.{major.minor}")
	need.Alternatives = []string{"python3"}
	if got := FindWingetFolder(local, need); got != dir {
		t.Fatalf("got %q, want %q", got, dir)
	}
}
