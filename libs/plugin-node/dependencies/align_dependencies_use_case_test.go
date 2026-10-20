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
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("package.json", `{"name": "acme", "devDependencies": {"jest": "^30.0.0"}}`)
	write("apps/web/package.json", `{"name": "web", "dependencies": {"axios": "^1.6.0"}, "devDependencies": {"jest": "^29.0.0"}}`)
	write("apps/api/package.json", `{"name": "api", "dependencies": {"axios": "^1.7.0"}}`)
	projects := []protocol.Project{{Name: "web", Root: "apps/web"}, {Name: "api", Root: "apps/api"}}
	res, err := Declared(ws, protocol.DeclaredParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	var changes []protocol.Rewrite
	sections := map[string]string{}
	for _, d := range res.Declarations {
		sections[d.Project+"/"+d.Name] = d.Section
		if d.Name == "axios" {
			changes = append(changes, protocol.Rewrite{Declaration: d, Spec: "^1.7.0"})
		}
	}
	if sections["web/axios"] != "dependencies" || sections["web/jest"] != "devDependencies" || sections["/jest"] != "devDependencies" {
		t.Fatalf("sections %v", sections)
	}
	out, err := Rewrite(ws, protocol.RewriteParams{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Files) != 2 {
		t.Fatalf("both projects change in one call: %v", out.Files)
	}
	for _, rel := range []string{"apps/web/package.json", "apps/api/package.json"} {
		raw, _ := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		if !strings.Contains(string(raw), `"axios": "^1.7.0"`) {
			t.Errorf("%s: %s", rel, raw)
		}
	}
	web, _ := os.ReadFile(filepath.Join(ws, "apps", "web", "package.json"))
	if !strings.Contains(string(web), `"jest": "^29.0.0"`) {
		t.Errorf("a dev tool is not touched: %s", web)
	}
}
