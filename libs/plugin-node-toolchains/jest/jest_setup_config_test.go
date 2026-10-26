package jest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func TestJestSetsUpTheStarterProject(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme"}`)
	write(t, ws, "apps/web/package.json", `{"name": "@acme/web", "scripts": {"build": "tsc"}}`)
	write(t, ws, "apps/web/src/greeting/greet.use-case.ts", "export {}\n")

	c := New("1.0.0")
	if caps := serve.CapabilitiesOf(c); !caps.Generate || caps.Discover {
		t.Fatalf("capabilities %+v", caps)
	}
	info, _ := c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if info.Kind != protocol.KindConvention || info.Languages[0] != "node" {
		t.Fatalf("describe %+v", info)
	}
	if _, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web"}); err != nil {
		t.Fatal(err)
	}
	spec := read(t, ws, "apps/web/src/greeting/greet.use-case.spec.ts")
	project := read(t, ws, "apps/web/package.json")
	if !strings.Contains(spec, "from '@jest/globals'") || !strings.Contains(project, `"test": "jest"`) || !strings.Contains(project, `"<rootDir>/src"`) {
		t.Fatalf("spec:\n%s\nproject:\n%s", spec, project)
	}
	if root := read(t, ws, "package.json"); !strings.Contains(root, `"ts-jest"`) {
		t.Fatalf("root:\n%s", root)
	}
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
