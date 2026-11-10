package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeIn(t *testing.T, ws, rel, body string) {
	t.Helper()
	p := filepath.Join(ws, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAPlannedLibraryCarriesTheGroupedFixAndARefusedOneItsReason(t *testing.T) {
	ws := t.TempDir()
	manifest := func(build string) string {
		return `{"name":"@a/lib","main":"./dist/index.js","types":"./dist/index.d.ts","exports":{".":{"types":"./dist/index.d.ts","default":"./dist/index.js"}},"scripts":{"build":"` + build + `"}}`
	}
	writeIn(t, ws, "packages/one/package.json", strings.Replace(manifest("tsc --outDir dist"), "lib", "one", 1))
	writeIn(t, ws, "packages/two/package.json", strings.Replace(manifest("rimraf dist && tsc --outDir dist"), "lib", "two", 1))
	writeIn(t, ws, "packages/one/src/index.ts", "")
	writeIn(t, ws, "packages/two/src/index.ts", "")
	writeIn(t, ws, "tsconfig.base.json", "{\"compilerOptions\": {\"strict\": true}}\n")
	builds := map[string]protocol.Task{"build": {Outputs: []string{"../../dist/packages/one"}}}
	one := protocol.Project{Name: "one", Root: "packages/one", Kind: protocol.ProjectPackage, Tasks: builds}
	two := protocol.Project{Name: "two", Root: "packages/two", Kind: protocol.ProjectPackage, Tasks: builds}
	plan := manifestfixes.PlanLibraryMoves(ws, []protocol.Project{one, two}, nil)

	f, ok := CheckLibraryBuildDir(ws, one, plan)
	if !ok || f.OK || f.Fix == nil || f.Fix.Description != manifestfixes.LibraryBuildDescription || !strings.Contains(f.Remedy, "doctor --fix --only") {
		t.Fatalf("a planned library carries the grouped fix: %+v", f)
	}
	f, _ = CheckLibraryBuildDir(ws, two, plan)
	if f.Fix != nil || !strings.Contains(f.Remedy, "rimraf") || !strings.Contains(f.Remedy, "main, types and exports in package.json as they are") {
		t.Fatalf("a refused library says why, and the by-hand hint keeps exports as published: %+v", f)
	}
	if strings.Contains(f.Remedy, "point package.json's main") {
		t.Errorf("the old hint is wrong for exports: %s", f.Remedy)
	}
}

func TestEntryFilesOfAPackageInTheNxLayoutAreJudgedInTheWorkspaceDist(t *testing.T) {
	ws := t.TempDir()
	writeIn(t, ws, "packages/kit/package.json", `{"name":"@a/kit","main":"./dist/index.js","types":"./dist/index.d.ts","scripts":{"build":"tsc --outDir ../../dist/packages/kit"}}`)
	writeIn(t, ws, "dist/packages/kit/index.js", "")
	p := protocol.Project{Name: "kit", Root: "packages/kit", Kind: protocol.ProjectPackage, Tasks: map[string]protocol.Task{"build": {}}}
	f, ok := CheckEntryFiles(ws, p)
	if !ok || f.OK || !strings.Contains(f.Detail, "dist/packages/kit/index.d.ts") {
		t.Fatalf("./dist/index.d.ts is dist/packages/kit/index.d.ts, which the build did not write: %+v", f)
	}
	writeIn(t, ws, "dist/packages/kit/index.d.ts", "")
	if f, _ = CheckEntryFiles(ws, p); !f.OK {
		t.Fatalf("both files are there: %+v", f)
	}
}
