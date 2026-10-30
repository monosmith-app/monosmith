package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-lua/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "cli", Directory: "apps/cli", Workspace: info, Stack: "tools"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "strings", Directory: "libs/strings", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(ws, "apps", "cli", "bin", "cli.lua")); err != nil {
		t.Error("an app has a script")
	}
	rockspec, _ := os.ReadFile(filepath.Join(ws, "libs", "strings", "strings-0.0.0-1.rockspec"))
	if strings.Contains(string(rockspec), "install") || strings.Contains(string(rockspec), "smithStack") {
		t.Errorf("a lib installs no script and records no stack:\n%s", rockspec)
	}
	found, err := discovery.Discover(ws, []string{"apps/cli/cli-0.0.0-1.rockspec", "libs/strings/strings-0.0.0-1.rockspec"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatal(err, found.Projects)
	}
	app, lib := found.Projects[0], found.Projects[1]
	if app.Kind != protocol.ProjectApp || app.Version != "0.0.0" || app.Stack != "tools" || lib.Kind != protocol.ProjectLib || lib.Tasks["test"].Command[0] != "busted" {
		t.Fatalf("%+v %+v", app, lib)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "x", Directory: "packages/x", Workspace: info}); err == nil {
		t.Error("a package is refused")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "cli", Directory: "apps/cli", Workspace: info}); err == nil {
		t.Error("an existing directory is refused")
	}
}
