package configedit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestSetTurnsOneVersionOnInTheDependenciesTable(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, workspace.FileName)
	_ = os.WriteFile(file, []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n"), 0o644)
	if err := Set(root, "dependencies.one_version", []string{"true"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(file)
	cfg, err := workspace.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Dependencies.OneVersion {
		t.Errorf("one_version was not written:\n%s", raw)
	}
}
