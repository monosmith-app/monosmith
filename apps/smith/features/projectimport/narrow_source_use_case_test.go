package projectimport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNarrowReadsTheSubFolderOfAPlainFolder(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src", "Lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := Narrow(Source{Arg: dir, Dir: dir}, "src/Lib", nil)
	if err != nil {
		t.Fatal(err)
	}
	if src.Path != "src/Lib" || src.Dir != filepath.Join(dir, "src", "Lib") {
		t.Errorf("got %+v", src)
	}
	if _, err := Narrow(Source{Arg: dir, Dir: dir}, "src/None", nil); err == nil || !strings.Contains(err.Error(), "has no folder src/None") {
		t.Errorf("a missing sub-folder gave %v", err)
	}
	if _, err := Narrow(Source{Arg: dir, Dir: dir}, "../out", nil); err == nil {
		t.Error("a path out of the source was accepted")
	}
}

func TestFolderNameComesFromThePathTheURLOrTheFolder(t *testing.T) {
	cases := map[string]Source{
		"Lib":     {Path: "src/Lib", Dir: "/tmp/x/source"},
		"scripts": {URL: "https://github.com/owner/scripts.git", Dir: "/tmp/x/source"},
		"tools":   {Dir: filepath.Join("a", "tools")},
	}
	for want, src := range cases {
		if got := FolderName(src); got != want {
			t.Errorf("FolderName(%+v) = %q, want %q", src, got, want)
		}
	}
}
