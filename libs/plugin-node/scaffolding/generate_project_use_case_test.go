package scaffolding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateWritesProjectAndWorkspaceOnce(t *testing.T) {

	ws := t.TempDir()
	params := protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "sdk", Directory: "packages/sdk", Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}}
	res, err := Generate(ws, params)
	if err != nil {
		t.Fatal(err)
	}
	if !res.InstallPending {
		t.Fatal("the host installs once: the plugin must report InstallPending")
	}
	joined := strings.Join(res.Files, ",")
	for _, want := range []string{"package.json", ".npmrc", "packages/sdk/package.json", "packages/sdk/src/greeting/greet.use-case.js", "packages/sdk/src/index.js"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %v", want, res.Files)
		}
	}
	if strings.Contains(joined, ".ts") || strings.Contains(joined, "eslint") {
		t.Fatalf("the node plugin alone writes plain JavaScript: %v", res.Files)
	}
	var pkg map[string]any
	raw, _ := os.ReadFile(filepath.Join(ws, "packages", "sdk", "package.json"))
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatalf("generated package.json is not JSON: %v", err)
	}
	if pkg["name"] != "@acme/sdk" || pkg["private"] != nil {
		t.Fatalf("package manifest: %v", pkg)
	}
	var root map[string]any
	raw, _ = os.ReadFile(filepath.Join(ws, "package.json"))
	_ = json.Unmarshal(raw, &root)
	if root["name"] != "acme-workspace" || root["workspaces"] == nil {
		t.Fatalf("root manifest: %v", root)
	}

	res, err = Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Workspace: params.Workspace})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if f == "tsconfig.base.json" || f == "package.json" {
			t.Fatalf("root file %s must not be rewritten", f)
		}
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Workspace: params.Workspace}); err == nil {
		t.Fatal("generating over an existing project must fail")
	}
}

func TestRootManifestKeepsUserEntries(t *testing.T) {
	Installer = func(string) error { return nil }
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name":"mine","private":true,"devDependencies":{"typescript":"5.0.0"},"scripts":{"x":"y"}}`), 0o644)
	set := JavaScript
	set.DevDependencies = map[string]string{"typescript": "~6.0.3", "@types/node": "^24.0.0"}
	if _, err := GenerateFrom(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui"}, set); err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	raw, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	_ = json.Unmarshal(raw, &root)
	dev := root["devDependencies"].(map[string]any)
	if root["name"] != "mine" || dev["typescript"] != "5.0.0" || dev["@types/node"] == nil || root["scripts"] == nil {
		t.Fatalf("user entries must survive: %v", root)
	}
}
