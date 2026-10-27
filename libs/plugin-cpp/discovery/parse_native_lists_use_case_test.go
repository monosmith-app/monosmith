package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSmithNativeMarksAProjectNative(t *testing.T) {
	for src, want := range map[string]bool{
		"project(a)\nset(SMITH_NATIVE ON)\n":       true,
		"project(a)\nset(SMITH_NATIVE \"TRUE\")\n": true,
		"project(a)\nset(SMITH_NATIVE OFF)\n":      false,
		"project(a)\n# set(SMITH_NATIVE ON)\n":     false,
		"project(a)\n":                             false,
	} {
		if got := ParseLists(src).Native; got != want {
			t.Errorf("%q: %v", src, got)
		}
	}
}

func TestDiscoverTagsANativeProject(t *testing.T) {
	ws := t.TempDir()
	for name, body := range map[string]string{"cli": "project(cli)\nset(SMITH_KIND app)\nset(SMITH_NATIVE ON)\nadd_executable(cli main.cpp)\n", "tool": "project(tool)\nset(SMITH_KIND app)\nadd_executable(tool main.cpp)\n"} {
		dir := filepath.Join(ws, "apps", name)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "CMakeLists.txt"), []byte(body), 0o644)
	}
	got, err := Discover(ws, []string{"apps/cli/CMakeLists.txt", "apps/tool/CMakeLists.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Projects[0].Tags) != 1 || got.Projects[0].Tags[0] != NativeTag || len(got.Projects[1].Tags) != 0 {
		t.Errorf("%+v", got.Projects)
	}
}
