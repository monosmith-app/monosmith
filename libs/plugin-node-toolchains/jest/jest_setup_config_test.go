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

func TestJestFollowsASiblingShapeThatCompilesWithSwc(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme"}`)
	write(t, ws, "apps/web/package.json", `{"name": "@acme/web"}`)

	c := New("1.0.0")
	info, _ := c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if len(info.Follows) != 2 || info.Follows[0] != "test.runner" || info.Follows[1] != "jest.config" {
		t.Fatalf("follows %v", info.Follows)
	}
	shape := map[string]string{"test.runner": "@swc/jest"}
	if _, err := c.Generate(protocol.GenerateParams{Name: "web", Directory: "apps/web", Shape: shape}); err != nil {
		t.Fatal(err)
	}
	project, root := read(t, ws, "apps/web/package.json"), read(t, ws, "package.json")
	if !strings.Contains(project, `"@swc/jest"`) || strings.Contains(project, `"ts-jest"`) {
		t.Fatalf("the transform is @swc/jest:\n%s", project)
	}
	if !strings.Contains(root, `"@swc/jest"`) || !strings.Contains(root, `"@swc/core"`) || strings.Contains(root, `"ts-jest"`) {
		t.Fatalf("the root has the swc packages and not ts-jest:\n%s", root)
	}
}

// Siblings that keep Jest in a jest.config.cts and a .spec.swcrc get the same
// two files, and the project's package.json has no "jest" field.
func TestJestFollowsSiblingsThatKeepTheConfigInFiles(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme"}`)
	write(t, ws, "packages/dep/package.json", `{"name": "@acme/dep"}`)

	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	shape := map[string]string{"test.runner": "@swc/jest", "jest.config": "file"}
	if _, err := c.Generate(protocol.GenerateParams{Name: "dep", Directory: "packages/dep", Shape: shape}); err != nil {
		t.Fatal(err)
	}
	if project := read(t, ws, "packages/dep/package.json"); strings.Contains(project, `"jest": {`) || !strings.Contains(project, `"test": "jest"`) {
		t.Errorf("the config is inline:\n%s", project)
	}
	config := read(t, ws, "packages/dep/jest.config.cts")
	for _, want := range []string{`.spec.swcrc`, `"displayName": "dep"`, `"@swc/jest", swcJestConfig`, `"roots"`} {
		if !strings.Contains(config, want) {
			t.Errorf("jest.config.cts lacks %s:\n%s", want, config)
		}
	}
	if swc := read(t, ws, "packages/dep/.spec.swcrc"); !strings.Contains(swc, `"syntax": "typescript"`) {
		t.Errorf(".spec.swcrc:\n%s", swc)
	}
	if root := read(t, ws, "package.json"); !strings.Contains(root, `"ts-node"`) {
		t.Errorf("the root lacks ts-node, which loads a .cts config:\n%s", root)
	}
}

func TestJestFileConfigWithTsJestIsPlainExports(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "package.json", `{"name": "acme"}`)
	write(t, ws, "packages/dep/package.json", `{"name": "@acme/dep"}`)
	c := New("1.0.0")
	_, _ = c.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := c.Generate(protocol.GenerateParams{Name: "dep", Directory: "packages/dep", Shape: map[string]string{"jest.config": "file"}}); err != nil {
		t.Fatal(err)
	}
	if config := read(t, ws, "packages/dep/jest.config.cts"); !strings.HasPrefix(config, "module.exports = {") || !strings.Contains(config, `"ts-jest"`) {
		t.Errorf("jest.config.cts:\n%s", config)
	}
}
