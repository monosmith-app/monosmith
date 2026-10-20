package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// One version per production dependency: the plugin names production and dev
// sections apart, and one rewrite call aligns a dependency in every project.
func TestRewriteAlignsAProductionDependencyAcrossProjects(t *testing.T) {
	ws := t.TempDir()
	files := map[string]string{
		"apps/api/pyproject.toml": "[project]\nname = \"api\"\ndependencies = [\"requests>=2.31\"]\n\n[project.optional-dependencies]\ndev = [\"pytest==8.2.0\"]\n",
		"apps/cli/pyproject.toml": "[project]\nname = \"cli\"\ndependencies = [\"requests>=2.32.3\"]\n\n[project.optional-dependencies]\ndev = [\"pytest==8.3.0\"]\n",
	}
	for rel, body := range files {
		_ = os.MkdirAll(filepath.Join(ws, filepath.Dir(filepath.FromSlash(rel))), 0o755)
		_ = os.WriteFile(filepath.Join(ws, filepath.FromSlash(rel)), []byte(body), 0o644)
	}
	res, err := Declared(ws, protocol.DeclaredParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "cli", Root: "apps/cli"}}})
	if err != nil {
		t.Fatal(err)
	}
	var changes []protocol.Rewrite
	for _, d := range res.Declarations {
		switch {
		case d.Name == "requests" && d.Section == "dependencies":
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: ">=2.32.3"})
		case d.Name == "pytest" && d.Section != "dev":
			t.Errorf("pytest is a dev tool: %+v", d)
		}
	}
	if len(changes) != 2 {
		t.Fatalf("requests is declared by both projects: %d", len(changes))
	}
	out, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Files) != 2 {
		t.Fatalf("both projects change in one call: %v", out.Files)
	}
	api, _ := os.ReadFile(filepath.Join(ws, "apps", "api", "pyproject.toml"))
	if !strings.Contains(string(api), `["requests>=2.32.3"]`) || !strings.Contains(string(api), `dev = ["pytest==8.2.0"]`) {
		t.Fatalf("aligned, dev tool untouched:\n%s", api)
	}
}
