package vite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A change of the workspace's tsconfig.base.json must reach every browser app
// the stack writes, through tsconfig.browser.json.
func TestEveryTsconfigOfTheStackReachesTheBase(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for name, choices := range map[string][]string{"react": {"react"}, "vue": {"vue"}, "svelte": {"svelte"}, "html": {"html"}} {
		t.Run(name, func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: name, Choices: choices, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(ws, tsconfigchain.Browser)); err != nil {
				t.Error("the stack writes tsconfig.browser.json")
			}
			if ok, err := tsconfigchain.ReachesBase(ws, filepath.Join(ws, "apps", "web", "tsconfig.json")); err != nil || !ok {
				t.Errorf("apps/web/tsconfig.json does not reach tsconfig.base.json: %v", err)
			}
			if bad, err := tsconfigchain.Unreached(ws); err != nil || len(bad) > 0 {
				t.Errorf("these tsconfig files do not reach tsconfig.base.json: %v (%v)", bad, err)
			}
		})
	}
}
