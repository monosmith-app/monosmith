package vite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheModuleAliasesFactReachesViteAndItsTsconfig(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	_, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "react",
		Facts:     map[string]any{protocol.FactModuleAliases: map[string]any{"@app": "src"}},
		Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}})
	if err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(ws, "apps", "web", rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if got := read("vite.config.ts"); !strings.Contains(got, `"@app": "/src",`) || !strings.Contains(got, "plugins: [react()]") {
		t.Errorf("vite.config.ts:\n%s", got)
	}
	if got := read("tsconfig.json"); !strings.Contains(got, `"@app/*": ["./src/*"]`) {
		t.Errorf("tsconfig.json:\n%s", got)
	}
}

func TestWithoutTheFactViteConfigHasNoAlias(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "react",
		Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "apps", "web", "vite.config.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "alias") {
		t.Errorf("vite.config.ts changed without facts:\n%s", raw)
	}
}
