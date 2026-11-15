package fullconversion

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMakeViewHoldsTheConvertedWorkspaceAndLeavesTheRealOneAlone(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	view, err := MakeView(root, []string{"package.json", "gone.txt"}, func(dir string) error {
		if err := os.Remove(filepath.Join(dir, "package.json")); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("schema = 1\n"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer view.Close()
	if view.Tree.Has("package.json") || !view.Tree.Has("smith.toml") {
		t.Fatalf("the view should hold what the conversion leaves: %v", view.Tree.Files())
	}
	if view.Tree.Modules != root {
		t.Errorf("installed packages are described from the real workspace, not %q", view.Tree.Modules)
	}
	if _, err := os.Stat(filepath.Join(root, "package.json")); err != nil {
		t.Error("the real workspace was changed")
	}
	if _, err := os.Stat(filepath.Join(root, "smith.toml")); err == nil {
		t.Error("smith.toml was written to the real workspace")
	}
	view.Close()
	if _, err := os.Stat(view.Dir); err == nil {
		t.Error("Close leaves the scratch copy behind")
	}
}
