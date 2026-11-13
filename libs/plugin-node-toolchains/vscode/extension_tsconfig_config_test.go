package vscode

import (
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A change of the workspace's tsconfig.base.json must reach the extension, as an
// app and as a package, through tsconfig.browser.json.
func TestEveryTsconfigOfTheStackReachesTheBase(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for _, kind := range []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectPackage} {
		t.Run(string(kind), func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			dir := string(kind) + "s/ext"
			if _, err := p.Generate(protocol.GenerateParams{Kind: kind, Name: "ext", Directory: dir, Stack: Stack, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
				t.Fatal(err)
			}
			if ok, err := tsconfigchain.ReachesBase(ws, filepath.Join(ws, filepath.FromSlash(dir), "tsconfig.json")); err != nil || !ok {
				t.Errorf("%s/tsconfig.json does not reach tsconfig.base.json: %v", dir, err)
			}
			if bad, err := tsconfigchain.Unreached(ws); err != nil || len(bad) > 0 {
				t.Errorf("these tsconfig files do not reach tsconfig.base.json: %v (%v)", bad, err)
			}
		})
	}
}
