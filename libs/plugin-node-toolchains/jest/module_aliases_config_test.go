package jest

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projectManifest(t *testing.T, facts map[string]any) string {
	t.Helper()
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme"}`)
	write(t, ws, "apps/web/package.json", `{"name": "@acme/web"}`)
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web", Facts: facts}); err != nil {
		t.Fatal(err)
	}
	return read(t, ws, "apps/web/package.json")
}

func TestTheModuleAliasesFactBecomesJestMappingsAndCompilerPaths(t *testing.T) {
	got := projectManifest(t, map[string]any{protocol.FactModuleAliases: map[string]any{"@app": "src"}})
	for _, want := range []string{`"^@app/(.*)$": "<rootDir>/src/$1"`, `"^@app$": "<rootDir>/src"`, `"@app/*"`, `"./src/*"`} {
		if !strings.Contains(got, want) {
			t.Errorf("package.json lacks %s:\n%s", want, got)
		}
	}
}

func TestWithoutTheFactJestKeepsItsDefaults(t *testing.T) {
	got := projectManifest(t, nil)
	if strings.Contains(got, "paths") || strings.Contains(got, "@app") {
		t.Errorf("jest config changed without facts:\n%s", got)
	}
}
