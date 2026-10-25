package vscode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "helper-tools", Directory: "packages/helper-tools", Stack: Stack, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "packages", "helper-tools", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Name, Publisher, Main string
		Engines               struct{ VSCode string }
		Smith                 struct{ Kind, Stack string }
		Scripts               map[string]string
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("package.json is JSON: %v\n%s", err, raw)
	}
	if m.Name != "helper-tools" || m.Publisher != "acme" || m.Main != "../../dist/packages/helper-tools/main.js" || m.Engines.VSCode != "^"+engine || m.Smith.Kind != "package" || m.Smith.Stack != Stack {
		t.Errorf("manifest %+v", m)
	}
	if _, err := os.Stat(filepath.Join(ws, "packages", "helper-tools", "src", "greeting", "greet.use-case.ts")); err != nil {
		t.Error(err)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "inhouse", Directory: "apps/inhouse", Stack: Stack, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(filepath.Join(ws, "apps", "inhouse", "package.json")); !json.Valid(raw) {
		t.Error("the app's manifest is JSON")
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x", Directory: "libs/x", Stack: Stack}); err == nil {
		t.Error("a lib is refused")
	}
}

func TestTheGeneratedReadmeNamesEveryWayToPublish(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "helper-tools", Directory: "packages/helper-tools", Stack: Stack, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "packages", "helper-tools", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"VSCE_PAT", "vsce_entra", "OVSX_PAT"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("README must mention %s:\n%s", want, raw)
		}
	}
}
