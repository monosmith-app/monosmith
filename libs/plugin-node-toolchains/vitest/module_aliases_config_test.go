package vitest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func generate(t *testing.T, viteConfig bool, facts map[string]any) (ws string) {
	t.Helper()
	ws = t.TempDir()
	files := map[string]string{"package.json": `{"name": "acme"}`, "libs/ui/package.json": `{"name": "@acme/ui"}`}
	if viteConfig {
		files["libs/ui/vite.config.ts"] = "export default {}\n"
	}
	for rel, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := c.Generate(protocol.GenerateParams{Name: "ui", Directory: "libs/ui", Facts: facts}); err != nil {
		t.Fatal(err)
	}
	return ws
}

var aliasFacts = map[string]any{protocol.FactModuleAliases: map[string]any{"@app": "src"}}

func TestTheModuleAliasesFactBecomesAVitestConfig(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(generate(t, false, aliasFacts), "libs", "ui", "vitest.config.ts"))
	if err != nil || !strings.Contains(string(raw), `'@app': '/src',`) {
		t.Fatalf("vitest.config.ts: %v\n%s", err, raw)
	}
}

func TestAViteConfigAlreadyCarriesTheAliasesForVitest(t *testing.T) {
	if _, err := os.Stat(filepath.Join(generate(t, true, aliasFacts), "libs", "ui", "vitest.config.ts")); err == nil {
		t.Error("a vitest.config.ts would hide the project's vite.config.ts")
	}
}

func TestWithoutTheFactVitestWritesNoConfig(t *testing.T) {
	if _, err := os.Stat(filepath.Join(generate(t, false, nil), "libs", "ui", "vitest.config.ts")); err == nil {
		t.Error("no facts, no vitest.config.ts")
	}
}
