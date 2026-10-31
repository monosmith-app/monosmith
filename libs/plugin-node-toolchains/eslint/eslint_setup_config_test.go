package eslint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestESLintAddsConfigAndLintScript(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name": "acme"}`), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "apps", "web", "package.json"), []byte(`{"name": "@acme/web"}`), 0o644)
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	res, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Files, ","); got != "apps/web/package.json,eslint.config.mjs,package.json" {
		t.Fatalf("files %s", got)
	}
	root, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	project, _ := os.ReadFile(filepath.Join(ws, "apps", "web", "package.json"))
	if !strings.Contains(string(root), `"@monosmith/eslint-config"`) || !strings.Contains(string(project), `"lint": "eslint ."`) {
		t.Fatalf("root:\n%s\nproject:\n%s", root, project)
	}
}

func TestESLintReactChoiceEnablesFrontEndRoles(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name": "acme"}`), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "apps", "web", "package.json"), []byte(`{"name": "@acme/web"}`), 0o644)
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web", Stack: "react-p5", Choices: []string{"react", "p5"}}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(ws, "eslint.config.mjs"))
	if !strings.Contains(string(got), "roles: ['hook', 'content', 'fixture', 'style', 'mock']") {
		t.Fatalf("config: %s", got)
	}
	if strings.Contains(configFor([]string{"jest"}), "roles:") {
		t.Fatal("a back-end stack must not get the front-end roles")
	}
}
