package angular

import (
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A change of the workspace's tsconfig.base.json must reach the Angular app,
// through tsconfig.browser.json.
func TestEveryTsconfigOfTheStackReachesTheBase(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "admin", Directory: "apps/admin", Stack: Name, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
		t.Fatal(err)
	}
	if ok, err := tsconfigchain.ReachesBase(ws, filepath.Join(ws, "apps", "admin", "tsconfig.json")); err != nil || !ok {
		t.Errorf("apps/admin/tsconfig.json does not reach tsconfig.base.json: %v", err)
	}
	if bad, err := tsconfigchain.Unreached(ws); err != nil || len(bad) > 0 {
		t.Errorf("these tsconfig files do not reach tsconfig.base.json: %v (%v)", bad, err)
	}
}
