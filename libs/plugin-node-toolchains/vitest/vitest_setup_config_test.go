package vitest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestVitestSetsUpTheStarterProject(t *testing.T) {
	ws := t.TempDir()
	for rel, body := range map[string]string{
		"package.json":                           `{"name": "acme"}`,
		"libs/ui/package.json":                   `{"name": "@acme/ui"}`,
		"libs/bare/package.json":                 `{"name": "@acme/bare"}`,
		"libs/ui/src/greeting/greet.use-case.ts": "export {}\n",
	} {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	res, err := c.Generate(protocol.GenerateParams{Name: "ui", Directory: "libs/ui"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Files, ","); got != "libs/ui/package.json,libs/ui/src/greeting/greet.use-case.spec.ts,package.json" {
		t.Fatalf("files %s", got)
	}
	spec, _ := os.ReadFile(filepath.Join(ws, "libs", "ui", "src", "greeting", "greet.use-case.spec.ts"))
	project, _ := os.ReadFile(filepath.Join(ws, "libs", "ui", "package.json"))
	if !strings.Contains(string(spec), "from 'vitest'") || !strings.Contains(string(project), `"test": "vitest run"`) {
		t.Fatalf("spec:\n%s\nproject:\n%s", spec, project)
	}

	// A project without the starter slice gets the script but no spec.
	res, err = c.Generate(protocol.GenerateParams{Name: "bare", Directory: "libs/bare"})
	if err != nil || strings.Join(res.Files, ",") != "libs/bare/package.json" {
		t.Fatalf("bare: %v %v", res.Files, err)
	}
}
