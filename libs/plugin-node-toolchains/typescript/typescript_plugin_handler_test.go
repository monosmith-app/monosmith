package typescript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func TestTypeScriptScaffoldsAndRecordsTheStack(t *testing.T) {
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	if caps := serve.CapabilitiesOf(p); !caps.Generate || caps.Discover {
		t.Fatalf("capabilities %+v", caps)
	}
	info, _ := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if info.Stacks[0].Scaffold != Name || strings.Join(info.Provides, ",") != "build,typecheck" {
		t.Fatalf("describe %+v", info)
	}
	res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "sdk", Directory: "packages/sdk", Stack: "node-ts", Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(res.Files, ",")
	for _, want := range []string{"tsconfig.base.json", ".npmrc", "packages/sdk/src/greeting/greet.use-case.ts", "packages/sdk/tsconfig.build.json"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %v", want, res.Files)
		}
	}
	manifest, _ := os.ReadFile(filepath.Join(ws, "packages", "sdk", "package.json"))
	for _, want := range []string{`"stack": "node-ts"`, `"build": "tsc -p tsconfig.build.json"`, `"name": "@acme/sdk"`} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("package.json lacks %s:\n%s", want, manifest)
		}
	}
	root, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	if !strings.Contains(string(root), `"typescript"`) {
		t.Fatalf("root devDependencies:\n%s", root)
	}
}

// A converted Nx workspace's base does not allow .ts specifiers, and the
// scaffold leaves a base it did not write alone: the project's own tsconfig
// sets what its starter code needs (#314).
func TestTypeScriptProjectAllowsItsOwnImportsUnderAnyBase(t *testing.T) {
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	converted := `{ "compilerOptions": { "module": "nodenext", "moduleResolution": "nodenext" } }` + "\n"
	if err := os.WriteFile(filepath.Join(ws, "tsconfig.base.json"), []byte(converted), 0o644); err != nil {
		t.Fatal(err)
	}
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "driver", Directory: "packages/driver", Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}}); err != nil {
		t.Fatal(err)
	}
	if base, _ := os.ReadFile(filepath.Join(ws, "tsconfig.base.json")); string(base) != converted {
		t.Errorf("the workspace's base was rewritten:\n%s", base)
	}
	project, _ := os.ReadFile(filepath.Join(ws, "packages", "driver", "tsconfig.json"))
	if !strings.Contains(string(project), `"rewriteRelativeImportExtensions": true`) {
		t.Errorf("the project's tsconfig does not allow its .ts imports:\n%s", project)
	}
	index, _ := os.ReadFile(filepath.Join(ws, "packages", "driver", "src", "index.ts"))
	if !strings.Contains(string(index), `'./greeting/index.ts'`) {
		t.Errorf("index.ts:\n%s", index)
	}
}
