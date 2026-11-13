package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestCINodeIsDecidedByToolpin checks the order: smith.toml, .nvmrc, then the
// default.
func TestCINodeIsDecidedByToolpin(t *testing.T) {
	ws := t.TempDir()
	n := &Node{root: ws}
	params := protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}}
	version := func() string {
		res, err := n.CI(params)
		if err != nil || len(res.Toolchains) == 0 {
			t.Fatalf("no toolchain: %v %+v", err, res)
		}
		return res.Toolchains[0].Version
	}
	if got := version(); got != "24" {
		t.Fatalf("default: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, ".nvmrc"), []byte("v22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := version(); got != "22" {
		t.Fatalf("native file: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("[toolchains]\nnode = \"20\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := version(); got != "20" {
		t.Fatalf("smith.toml: got %q", got)
	}
}
