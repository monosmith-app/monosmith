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
	for _, want := range []string{`"stack": "node-ts"`, `"build": "tsc -p tsconfig.build.json --outDir ../../dist/packages/sdk"`, `"name": "@acme/sdk"`} {
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

func generated(t *testing.T, kind protocol.ProjectKind, shape map[string]string) (project, root string) {
	t.Helper()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	info, _ := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if len(info.Follows) != 5 {
		t.Fatalf("follows %v", info.Follows)
	}
	params := protocol.GenerateParams{Kind: kind, Name: "dep", Directory: "packages/dep", Shape: shape, Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}}
	if _, err := p.Generate(params); err != nil {
		t.Fatal(err)
	}
	manifest, _ := os.ReadFile(filepath.Join(ws, "packages", "dep", "package.json"))
	rootManifest, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	return string(manifest), string(rootManifest)
}

func TestTypeScriptLibFollowsSiblingsThatBuildWithEsbuildIntoDist(t *testing.T) {
	project, root := generated(t, protocol.ProjectLib, map[string]string{"build.tool": "esbuild", "exports.target": "dist"})
	for _, want := range []string{`"build": "esbuild src/index.ts`, `--outdir=../../dist/packages/dep`, `"main": "../../dist/packages/dep/index.js"`, `"types": "../../dist/packages/dep/index.d.ts"`} {
		if !strings.Contains(project, want) {
			t.Errorf("package.json lacks %s:\n%s", want, project)
		}
	}
	if strings.Contains(project, `"exports"`) {
		t.Errorf("a lib that exports its build has no exports field:\n%s", project)
	}
	if !strings.Contains(root, `"esbuild"`) {
		t.Errorf("the root does not pin esbuild:\n%s", root)
	}
}

func TestTypeScriptPackageAndAppFollowTheBuildTool(t *testing.T) {
	pkg, _ := generated(t, protocol.ProjectPackage, map[string]string{"build.tool": "esbuild", "exports.target": "dist"})
	if !strings.Contains(pkg, `esbuild src/index.ts`) || !strings.Contains(pkg, `"types": "../../dist/packages/dep/index.d.ts"`) {
		t.Errorf("package:\n%s", pkg)
	}
	app, _ := generated(t, protocol.ProjectApp, map[string]string{"build.tool": "esbuild"})
	if !strings.Contains(app, `esbuild src/main.ts`) {
		t.Errorf("app:\n%s", app)
	}
}

func TestTypeScriptLibWithoutShapeKeepsTheDefault(t *testing.T) {
	project, root := generated(t, protocol.ProjectLib, nil)
	if !strings.Contains(project, `"exports": "./src/index.ts"`) || strings.Contains(project, `"build"`) {
		t.Errorf("the default lib exports its source and has no build:\n%s", project)
	}
	if strings.Contains(root, `"esbuild"`) {
		t.Errorf("the root pins esbuild without being asked:\n%s", root)
	}
}

func generatedFiles(t *testing.T, kind protocol.ProjectKind, shape map[string]string) (ws string) {
	t.Helper()
	scaffolding.Installer = func(string) error { return nil }
	ws = t.TempDir()
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	params := protocol.GenerateParams{Kind: kind, Name: "dep", Directory: "packages/dep", Shape: shape, Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}}
	if _, err := p.Generate(params); err != nil {
		t.Fatal(err)
	}
	return ws
}

func read(t *testing.T, ws, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return string(raw)
}

// Siblings with tsconfig.json, tsconfig.lib.json and tsconfig.spec.json give
// the new package the same three files, and the scripts read the right one.
func TestTypeScriptPackageFollowsTheTsconfigSetOfSiblings(t *testing.T) {
	ws := generatedFiles(t, protocol.ProjectPackage, map[string]string{
		"build.tool": "esbuild", "exports.target": "dist",
		"tsconfig.files": "tsconfig.json,tsconfig.lib.json,tsconfig.spec.json",
	})
	if _, err := os.Stat(filepath.Join(ws, "packages", "dep", "tsconfig.build.json")); err == nil {
		t.Errorf("tsconfig.build.json written although the siblings have none")
	}
	solution := read(t, ws, "packages/dep/tsconfig.json")
	for _, want := range []string{`"files": []`, `./tsconfig.lib.json`, `./tsconfig.spec.json`} {
		if !strings.Contains(solution, want) {
			t.Errorf("tsconfig.json lacks %s:\n%s", want, solution)
		}
	}
	if lib := read(t, ws, "packages/dep/tsconfig.lib.json"); !strings.Contains(lib, `"rootDir": "src"`) || !strings.Contains(lib, `.spec.ts`) {
		t.Errorf("tsconfig.lib.json:\n%s", lib)
	}
	if spec := read(t, ws, "packages/dep/tsconfig.spec.json"); !strings.Contains(spec, `"include": ["src/**/*.spec.ts"]`) {
		t.Errorf("tsconfig.spec.json:\n%s", spec)
	}
	manifest := read(t, ws, "packages/dep/package.json")
	for _, want := range []string{`tsc -p tsconfig.lib.json --declaration`, `"typecheck": "tsc -p tsconfig.lib.json --noEmit"`} {
		if !strings.Contains(manifest, want) {
			t.Errorf("package.json lacks %s:\n%s", want, manifest)
		}
	}
}

// Siblings that build into their own dist folder and export a map get the same.
func TestTypeScriptPackageFollowsSiblingsThatBuildIntoTheirOwnDist(t *testing.T) {
	manifest, _ := generated(t, protocol.ProjectPackage, map[string]string{
		"build.tool": "esbuild", "exports.target": "dist", "build.output": "project", "exports.map": "map",
	})
	for _, want := range []string{
		`--outdir=dist &&`, `--declarationDir dist"`, `"main": "./dist/index.js"`,
		`"import": "./dist/index.js"`, `"files": ["dist"]`, `"./package.json": "./package.json"`,
	} {
		if !strings.Contains(manifest, want) {
			t.Errorf("package.json lacks %s:\n%s", want, manifest)
		}
	}
	if strings.Contains(manifest, "../../dist") || strings.Contains(manifest, "--outDir") {
		t.Errorf("the build leaves the project folder:\n%s", manifest)
	}
}
