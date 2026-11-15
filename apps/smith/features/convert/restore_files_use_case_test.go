package convert

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotPutsBackChangedDeletedAndCreatedFiles(t *testing.T) {
	root := t.TempDir()
	write := func(rel, text string) {
		t.Helper()
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", "{}")
	write("nx.json", "{}")
	write("apps/a/project.json", "{}")
	write(".nx/cache/big", "kept out")

	snap := newSnapshot(root)
	for _, rel := range []string{"package.json", "nx.json", "apps", "smith.toml", ".nx"} {
		snap.remember(rel)
	}
	write("package.json", `{"changed":true}`)
	_ = os.Remove(filepath.Join(root, "nx.json"))
	_ = os.RemoveAll(filepath.Join(root, "apps"))
	write("smith.toml", "new")

	if err := snap.restore(); err != nil {
		t.Fatal(err)
	}
	for rel, want := range map[string]string{"package.json": "{}", "nx.json": "{}", "apps/a/project.json": "{}"} {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", rel, got, err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "smith.toml")); !os.IsNotExist(err) {
		t.Errorf("smith.toml did not exist before and must go: %v", err)
	}
	if _, ok := snap.saved[".nx/cache/big"]; ok {
		t.Error("the .nx cache is not remembered")
	}
}
