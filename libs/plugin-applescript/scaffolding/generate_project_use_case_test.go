package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-applescript/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "notes", Directory: "apps/notes", Stack: "mac"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "dialogs", Directory: "libs/dialogs"}); err != nil {
		t.Fatal(err)
	}
	app, _ := os.ReadFile(filepath.Join(ws, "apps", "notes", "main.applescript"))
	lib, _ := os.ReadFile(filepath.Join(ws, "libs", "dialogs", "main.applescript"))
	if !strings.Contains(string(app), "on run") || !strings.Contains(string(lib), "on greet(who)") || strings.Contains(string(lib), "smithStack") {
		t.Errorf("app:\n%s\nlib:\n%s", app, lib)
	}
	found, err := discovery.Discover(ws, []string{"apps/notes/main.applescript", "libs/dialogs/main.applescript"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatal(err, found.Projects)
	}
	if p := found.Projects[0]; p.Kind != protocol.ProjectApp || p.Version != "0.0.0" || p.Stack != "mac" || found.Projects[1].Kind != protocol.ProjectLib {
		t.Fatalf("%+v", found.Projects)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "x", Directory: "packages/x"}); err == nil {
		t.Error("a package is refused")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "notes", Directory: "apps/notes"}); err == nil {
		t.Error("an existing directory is refused")
	}
}
