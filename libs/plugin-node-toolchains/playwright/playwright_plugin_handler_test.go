package playwright

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func generated(t *testing.T, convention bool) (*Playwright, string, protocol.GenerateResult) {
	t.Helper()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "driver")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name":"ws","private":true,"workspaces":["packages/*"]}`+"\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"@acme/driver","version":"0.0.0"}`+"\n"), 0o644)
	p := New("1.0.0")
	if _, err := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws}); err != nil {
		t.Fatal(err)
	}
	res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "driver", Directory: "packages/driver", Convention: convention})
	if err != nil {
		t.Fatal(err)
	}
	return p, ws, res
}

// As a workspace convention, playwright runs the e2e of projects that have a
// suite; it never gives a new project one (#314).
func TestGenerateAsAConventionAddsNoSuite(t *testing.T) {
	_, ws, res := generated(t, true)
	if len(res.Files) != 0 {
		t.Errorf("wrote %v", res.Files)
	}
	if _, err := os.Stat(filepath.Join(ws, "packages", "driver", "playwright.config.ts")); err == nil {
		t.Error("a convention scaffolded a Playwright suite")
	}
}

// A stack that names playwright (node-vitest-playwright) still gets its suite.
func TestGenerateForTheStackAddsTheSuite(t *testing.T) {
	_, ws, _ := generated(t, false)
	for _, f := range []string{"playwright.config.ts", "e2e/smoke.e2e.ts"} {
		if _, err := os.Stat(filepath.Join(ws, "packages", "driver", filepath.FromSlash(f))); err != nil {
			t.Errorf("no %s: %v", f, err)
		}
	}
}
