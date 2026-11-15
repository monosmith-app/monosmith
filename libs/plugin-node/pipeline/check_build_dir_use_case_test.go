package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckBuildDir(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	builds := map[string]protocol.Task{"build": {}}
	write("apps/web/tsconfig.json", "{\n  // ours\n  \"compilerOptions\": { \"outDir\": \"../../dist/apps/web\" }\n}\n")
	if f, ok := CheckBuildDir(ws, protocol.Project{Name: "web", Root: "apps/web", Kind: protocol.ProjectApp, Tasks: builds}); !ok || !f.OK {
		t.Fatalf("web: %+v %v", f, ok)
	}
	write("apps/old/tsconfig.build.json", `{"compilerOptions": {"outDir": "dist"}}`)
	f, ok := CheckBuildDir(ws, protocol.Project{Name: "old", Root: "apps/old", Kind: protocol.ProjectApp, Tasks: builds})
	if !ok || f.OK || f.Remedy == "" {
		t.Fatalf("old: %+v", f)
	}
	if f.Fix != nil {
		t.Fatalf("a package without a package.json has nothing to edit: %+v", f.Fix)
	}
	if _, ok := CheckBuildDir(ws, protocol.Project{Name: "ui", Root: "libs/ui"}); ok {
		t.Fatal("a project without a build has nothing to check")
	}
}

func TestAMisplacedBuildOffersAQuestionOrRefusesWithTheReason(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "kit")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"@a/kit","private":true,"main":"./dist/index.js","scripts":{"build":"esbuild src/index.ts --outdir=dist"}}`), 0o644)
	p := protocol.Project{Name: "kit", Root: "packages/kit", Kind: protocol.ProjectApp, Exports: []string{"@a/kit"}, Tasks: map[string]protocol.Task{"build": {Outputs: []string{"../../dist/packages/kit"}}}}

	f, ok := CheckBuildDir(ws, p)
	if !ok || f.OK || f.Fix == nil || len(f.Fix.Choices) != 1 || len(f.Fix.Files) != 0 {
		t.Fatalf("a question, nothing written before the answer: %+v", f)
	}
	if !strings.Contains(f.Detail, "packages/kit/dist") || !strings.Contains(f.Detail, "caches dist/packages/kit instead") {
		t.Errorf("the finding names where it builds and what the task caches: %s", f.Detail)
	}

	f, _ = CheckBuildDir(ws, p, "web")
	if f.Fix != nil || !strings.Contains(f.Remedy, "web import it by name") {
		t.Fatalf("an imported package is refused with the reason: %+v", f)
	}
	if f.Warning {
		t.Errorf("a cache that restores nothing is a failure, even when the move is refused: %+v", f)
	}
	p.Tasks = map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}
	f, _ = CheckBuildDir(ws, p, "web")
	if f.OK || !f.Warning || f.Fix != nil || !strings.Contains(f.Remedy, "To move it by hand") {
		t.Errorf("a refused move of a build that works is a warning with the reason and the manual step: %+v", f)
	}
	if got := dependentsOf([]protocol.Project{p, {Name: "web", Imports: []string{"@a/kit"}}, {Name: "other", Imports: []string{"x"}}}, p); len(got) != 1 || got[0] != "web" {
		t.Errorf("dependents: %v", got)
	}
}

func TestTheBuildScriptsOutDirFlagIsTheOutputFolder(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs", "kit")
	_ = os.MkdirAll(dir, 0o755)
	builds := map[string]protocol.Task{"build": {}}
	p := protocol.Project{Name: "kit", Root: "libs/kit", Kind: protocol.ProjectApp, Tasks: builds}
	_ = os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"compilerOptions": {"outDir": "dist"}}`), 0o644)

	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"tsc -p tsconfig.build.json --outDir ../../dist/libs/kit"}}`), 0o644)
	if f, ok := CheckBuildDir(ws, p); !ok || !f.OK {
		t.Fatalf("the flag wins over the tsconfig: %+v %v", f, ok)
	}
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"tsc --outDir=lib"}}`), 0o644)
	if f, ok := CheckBuildDir(ws, p); !ok || f.OK || f.Remedy == "" {
		t.Fatalf("a flag inside the source tree is flagged: %+v %v", f, ok)
	}
}

func TestALibraryBuildsInsideItsOwnFolder(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	builds := map[string]protocol.Task{"build": {}}
	write("packages/kit/tsconfig.json", `{"compilerOptions": {"outDir": "dist"}}`)
	write("packages/kit/package.json", `{"name": "@a/kit", "main": "./dist/index.js", "scripts": {"build": "tsc"}}`)
	kit := protocol.Project{Name: "kit", Root: "packages/kit", Kind: protocol.ProjectLib, PackageName: "@a/kit", Tasks: builds}
	if f, ok := CheckBuildDir(ws, kit); !ok || !f.OK {
		t.Errorf("a library that builds into its own dist is right: %+v", f)
	}

	// The layout that broke the scoped rename: main leads out of the package.
	write("libs/models/tsconfig.json", "{\n  \"compilerOptions\": {\n    \"outDir\": \"../../dist/libs/models\",\n    \"strict\": true\n  }\n}\n")
	write("libs/models/package.json", `{"name": "@q/models", "main": "../../dist/libs/models/index.js", "types": "../../dist/libs/models/index.d.ts", "scripts": {"build": "tsc"}}`)
	models := protocol.Project{Name: "models", Root: "libs/models", Kind: protocol.ProjectLib, PackageName: "@q/models", Tasks: builds}
	f, ok := CheckBuildDir(ws, models)
	if !ok || f.OK || f.Warning || f.Fix == nil || !strings.Contains(f.Detail, "node_modules/@q/models") {
		t.Fatalf("a library that builds outside its folder fails with a fix: %+v", f)
	}
	edits := map[string]string{}
	for _, e := range f.Fix.Files {
		edits[e.Path] = e.Content
	}
	if !strings.Contains(edits["libs/models/tsconfig.json"], `"outDir": "./dist"`) || !strings.Contains(edits["libs/models/tsconfig.json"], `"strict": true`) {
		t.Errorf("the tsconfig names ./dist and keeps the rest:\n%s", edits["libs/models/tsconfig.json"])
	}
	if pkg := edits["libs/models/package.json"]; !strings.Contains(pkg, `"main": "./dist/index.js"`) || !strings.Contains(pkg, `"types": "./dist/index.d.ts"`) {
		t.Errorf("main and types follow the build into the package:\n%s", pkg)
	}

	write("apps/tool/tsconfig.json", `{"compilerOptions": {"outDir": "dist"}}`)
	write("apps/tool/package.json", `{"name": "tool", "scripts": {"build": "tsc"}}`)
	app, _ := CheckBuildDir(ws, protocol.Project{Name: "tool", Root: "apps/tool", Kind: protocol.ProjectApp, Tasks: builds})
	if app.OK || app.Warning {
		t.Errorf("an app still builds into the workspace's dist, and one smith can move is a failure: %+v", app)
	}
}
