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
	if f, ok := CheckBuildDir(ws, protocol.Project{Name: "web", Root: "apps/web", Tasks: builds}); !ok || !f.OK {
		t.Fatalf("web: %+v %v", f, ok)
	}
	write("apps/old/tsconfig.build.json", `{"compilerOptions": {"outDir": "dist"}}`)
	f, ok := CheckBuildDir(ws, protocol.Project{Name: "old", Root: "apps/old", Tasks: builds})
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
	p := protocol.Project{Name: "kit", Root: "packages/kit", Exports: []string{"@a/kit"}, Tasks: map[string]protocol.Task{"build": {Outputs: []string{"../../dist/packages/kit"}}}}

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
	if got := dependentsOf([]protocol.Project{p, {Name: "web", Imports: []string{"@a/kit"}}, {Name: "other", Imports: []string{"x"}}}, p); len(got) != 1 || got[0] != "web" {
		t.Errorf("dependents: %v", got)
	}
}

func TestTheBuildScriptsOutDirFlagIsTheOutputFolder(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs", "kit")
	_ = os.MkdirAll(dir, 0o755)
	builds := map[string]protocol.Task{"build": {}}
	p := protocol.Project{Name: "kit", Root: "libs/kit", Tasks: builds}
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
