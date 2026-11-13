package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestCIPythonIsDecidedByToolpin checks the order: smith.toml, .python-version,
// then the default.
func TestCIPythonIsDecidedByToolpin(t *testing.T) {
	ws := t.TempDir()
	params := protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}}
	if err := os.WriteFile(filepath.Join(ws, ".python-version"), []byte("3.13\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := CI(ws, params).Toolchains[0].Version; got != "3.13" {
		t.Fatalf("native file: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\npython = \"3.14\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := CI(ws, params).Toolchains[0].Version; got != "3.14" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
