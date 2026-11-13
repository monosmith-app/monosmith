package eslint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
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

// A workspace on an ESLint older than the config's peer gets the peer range:
// that ESLint cannot load @monosmith/eslint-config. A newer 10.x stays.
func TestESLintRaisesAnESLintTheConfigCannotRunWith(t *testing.T) {
	for _, tc := range []struct{ has, want string }{
		{`"eslint": "^9.39.3"`, `"eslint": "` + eslintpeer.Range + `"`},
		{`"eslint": "^10.12.0"`, `"eslint": "^10.12.0"`},
	} {
		ws := t.TempDir()
		_ = os.MkdirAll(filepath.Join(ws, "apps", "web"), 0o755)
		_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name": "acme", "devDependencies": {`+tc.has+`}}`), 0o644)
		_ = os.WriteFile(filepath.Join(ws, "apps", "web", "package.json"), []byte(`{"name": "@acme/web"}`), 0o644)
		c := New("1.0.0")
		_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
		if _, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web"}); err != nil {
			t.Fatal(err)
		}
		root, _ := os.ReadFile(filepath.Join(ws, "package.json"))
		if !strings.Contains(string(root), tc.want) || !strings.Contains(string(root), `"`+eslintpeer.ConfigPackage+`"`) {
			t.Errorf("with %s, want %s:\n%s", tc.has, tc.want, root)
		}
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
