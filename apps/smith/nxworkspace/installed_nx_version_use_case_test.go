package nxworkspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstalledNxVersionReadsTheInstalledPackageForARange(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "node_modules", "nx")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"version":"22.3.1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for rng, want := range map[string]string{"*": "22.3.1", "^22.0.0": "22.3.1", "catalog:": "22.3.1", "21.0.0": "21.0.0"} {
		if got := installedNxVersion(root, rng); got != want {
			t.Errorf("range %q: got %q, want %q", rng, got, want)
		}
	}
	if got := installedNxVersion(t.TempDir(), "*"); got != "*" {
		t.Errorf("without node_modules the range stays: %q", got)
	}
}
