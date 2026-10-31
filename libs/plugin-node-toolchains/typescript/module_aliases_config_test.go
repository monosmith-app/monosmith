package typescript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func tsconfigOf(t *testing.T, facts map[string]any) string {
	t.Helper()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "sdk", Directory: "libs/sdk", Stack: "node-ts", Facts: facts,
		Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "libs", "sdk", "tsconfig.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestTheModuleAliasesFactBecomesTsconfigPaths(t *testing.T) {
	got := tsconfigOf(t, map[string]any{protocol.FactModuleAliases: map[string]any{"@app": "src"}})
	for _, want := range []string{`"paths": {`, `"@app": ["./src"]`, `"@app/*": ["./src/*"]`} {
		if !strings.Contains(got, want) {
			t.Errorf("tsconfig lacks %s:\n%s", want, got)
		}
	}
}

func TestThePluginPublishesItsSourceRootAsAFact(t *testing.T) {
	res, err := New("1.0.0").Contribute(protocol.ContributeParams{Projects: []protocol.Project{{Name: "web"}}})
	if err != nil {
		t.Fatal(err)
	}
	facts := res.Projects["web"].Facts
	if err := protocol.ValidateFacts(facts); err != nil || facts[protocol.FactSourceRoots] == nil {
		t.Fatalf("facts %v: %v", facts, err)
	}
}

func TestWithoutTheFactTsconfigIsAsBefore(t *testing.T) {
	got := tsconfigOf(t, nil)
	want := "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"compilerOptions\": {\n    \"rootDir\": \"src\",\n    \"rewriteRelativeImportExtensions\": true,\n    \"types\": [\"node\"]\n  },\n  \"include\": [\"src\"]\n}\n"
	if strings.ReplaceAll(got, "\r\n", "\n") != want {
		t.Errorf("tsconfig changed without facts:\n%s", got)
	}
}
