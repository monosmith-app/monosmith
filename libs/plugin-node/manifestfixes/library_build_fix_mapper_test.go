package manifestfixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const libraryBuildScript = "esbuild src/index.ts --bundle --outfile=dist/index.esm.js && tsc -p tsconfig.lib.json --emitDeclarationOnly --declaration --declarationDir dist/src && smith declaration-specifiers dist/src"

const libraryTSConfig = "{\n  \"compilerOptions\": {\n    \"rootDir\": \"src\",\n    \"outDir\": \"dist\"\n  }\n}\n"

func libraryManifest(name, buildScript string) string {
	return `{
  "name": "@a/` + name + `",
  "files": ["dist", "!**/*.map"],
  "main": "./dist/index.esm.js",
  "types": "./dist/src/index.d.ts",
  "exports": {
    "./package.json": "./package.json",
    ".": { "types": "./dist/src/index.d.ts", "default": "./dist/index.esm.js" }
  },
  "scripts": { "build": "` + buildScript + `", "test": "jest" }
}
`
}

// libraryWorkspace writes packages one and two, and the workspace files, into a
// temporary folder.
func libraryWorkspace(t *testing.T, extra map[string]string) (string, []protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{
		"tsconfig.base.json":              "{\n  \"compilerOptions\": {\n    \"strict\": true\n  }\n}\n",
		"jest.preset.js":                  "const nx = require('@monosmith/jest-preset')\nmodule.exports = { ...nx }\n",
		"packages/one/package.json":       libraryManifest("one", libraryBuildScript),
		"packages/one/tsconfig.lib.json":  libraryTSConfig,
		"packages/one/src/index.ts":       "export const one = 1\n",
		"packages/two/package.json":       libraryManifest("two", libraryBuildScript),
		"packages/two/tsconfig.lib.json":  libraryTSConfig,
		"packages/two/src/index.ts":       "export const two = 2\n",
		"smith.toml":                      "[[groups]]\nselect = [\"lang:node\"]\n\n[groups.tasks.build]\noutputs = [\"dist\"]\n\n[groups.tasks.test]\nenv = { A = \"b\" }\n",
		"packages/two/tsconfig.spec.json": "{\"compilerOptions\": {\"outDir\": \"./out-tsc/jest\"}}\n",
	}
	for rel, body := range extra {
		files[rel] = body
	}
	for rel, body := range files {
		abs := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(abs), 0o755)
		_ = os.WriteFile(abs, []byte(body), 0o644)
	}
	task := map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}
	return ws, []protocol.Project{
		{Name: "one", Root: "packages/one", Kind: protocol.ProjectPackage, Tasks: task},
		{Name: "two", Root: "packages/two", Kind: protocol.ProjectPackage, Tasks: task},
	}
}

func applyFiles(t *testing.T, ws string, files []protocol.FileEdit) {
	t.Helper()
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(ws, filepath.FromSlash(f.Path)), []byte(f.Content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func contentOf(files []protocol.FileEdit, path string) string {
	for _, f := range files {
		if f.Path == path {
			return f.Content
		}
	}
	return ""
}

func TestPlanLibraryMovesGroupsEveryPackageIntoOneQuestion(t *testing.T) {
	ws, projects := libraryWorkspace(t, nil)
	plan := PlanLibraryMoves(ws, projects, map[string][]string{"one": {"two"}})
	if plan.Fix == nil || len(plan.Fix.Choices) != 1 || len(plan.Fix.Files) != 0 {
		t.Fatalf("one question, nothing written before the answer: %+v", plan.Fix)
	}
	if plan.Fix.Description != LibraryBuildDescription || strings.Join(plan.Packages, ",") != "one,two" {
		t.Fatalf("%q %v", plan.Fix.Description, plan.Packages)
	}
	// "two" has "exports" and no dependents: it is covered because exports cannot name ../../dist.
	move := plan.Fix.Choices[0].Options[0].Fix
	if len(plan.Fix.Choices[0].Options) != 2 || len(move.Choices) != 0 {
		t.Fatalf("move or leave; Jest needs no question with @monosmith/jest-preset: %+v", plan.Fix.Choices)
	}
	manifest := contentOf(move.Files, "packages/one/package.json")
	if !strings.Contains(manifest, "--outfile=../../dist/packages/one/index.esm.js") || !strings.Contains(manifest, "--declarationDir ../../dist/packages/one/src") || !strings.Contains(manifest, `"main": "./dist/index.esm.js"`) {
		t.Errorf("the build script moves, main stays the published form:\n%s", manifest)
	}
	config := contentOf(move.Files, "packages/one/tsconfig.lib.json")
	if !strings.Contains(config, `"outDir": "../../dist/packages/one"`) || !strings.Contains(config, `"declarationDir": "../../dist/packages/one/src"`) {
		t.Errorf("the tsconfig moves and names the declaration folder of the script:\n%s", config)
	}
	if contentOf(move.Files, "packages/two/tsconfig.spec.json") != "" {
		t.Error("a tsconfig with no output in the old folder is not edited")
	}
	base := contentOf(move.Files, "tsconfig.base.json")
	if !strings.Contains(base, `"@a/one": ["./packages/one/src/index.ts"]`) || !strings.Contains(base, `"@a/two": ["./packages/two/src/index.ts"]`) {
		t.Errorf("paths map each name to its source:\n%s", base)
	}
	if toml := contentOf(move.Files, "smith.toml"); strings.Contains(toml, "outputs") || !strings.Contains(toml, "[groups.tasks.test]") {
		t.Errorf("the group's outputs line goes, nothing else:\n%s", toml)
	}
}

func TestPlanLibraryMovesIsIdempotent(t *testing.T) {
	ws, projects := libraryWorkspace(t, nil)
	plan := PlanLibraryMoves(ws, projects, map[string][]string{"one": {"two"}})
	applyFiles(t, ws, plan.Fix.Choices[0].Options[0].Fix.Files)
	again := PlanLibraryMoves(ws, projects, map[string][]string{"one": {"two"}})
	if again.Fix != nil || len(again.Packages) != 0 || len(again.Refusals) != 0 {
		t.Fatalf("a moved workspace has nothing left to move: %+v", again)
	}
}

func TestPlanLibraryMovesLeavesAPackageItCannotReadWithItsReason(t *testing.T) {
	ws, projects := libraryWorkspace(t, map[string]string{
		"packages/two/package.json": libraryManifest("two", "rimraf dist && "+libraryBuildScript),
	})
	plan := PlanLibraryMoves(ws, projects, nil)
	if reason := plan.Refusals["two"]; !strings.Contains(reason, "rimraf") {
		t.Fatalf("the reason names the command: %v", plan.Refusals)
	}
	if !plan.Moves("one") || plan.Moves("two") {
		t.Fatalf("only the readable one moves: %v", plan.Packages)
	}
	if strings.Contains(plan.Fix.Choices[0].Question, "outputs") && contentOf(plan.Fix.Choices[0].Options[0].Fix.Files, "smith.toml") != "" {
		t.Error("with a package left in dist, the group's outputs stay")
	}
}

func TestPlanLibraryMovesRefusesAnAliasMappedElsewhere(t *testing.T) {
	ws, projects := libraryWorkspace(t, map[string]string{
		"tsconfig.base.json": "{\n  \"compilerOptions\": {\n    \"paths\": { \"@a/one\": [\"vendor/one.ts\"] }\n  }\n}\n",
	})
	plan := PlanLibraryMoves(ws, projects, nil)
	if !strings.Contains(plan.Refusals["one"], "vendor/one.ts") || !plan.Moves("two") {
		t.Fatalf("%+v", plan)
	}
}

func TestPlanLibraryMovesAsksHowJestFindsThePackages(t *testing.T) {
	ws, projects := libraryWorkspace(t, map[string]string{"jest.preset.js": "module.exports = { testEnvironment: 'node' }\n"})
	plan := PlanLibraryMoves(ws, projects, nil)
	move := plan.Fix.Choices[0].Options[0].Fix
	if len(move.Choices) != 1 || len(move.Choices[0].Options) != 2 {
		t.Fatalf("a preset that reads no tsconfig paths gets a question: %+v", move.Choices)
	}
	mapper := move.Choices[0].Options[0].Fix.Files[0].Content
	if !strings.Contains(mapper, `'^@a/one$': require('node:path').join(__dirname, 'packages/one/src/index.ts')`) || !strings.HasPrefix(mapper, "module.exports = { testEnvironment: 'node' }\n") {
		t.Errorf("one mapper entry per package, the preset's own text first:\n%s", mapper)
	}
}

func TestPlanLibraryMovesNeedsATSConfigBase(t *testing.T) {
	ws, projects := libraryWorkspace(t, nil)
	_ = os.Remove(filepath.Join(ws, "tsconfig.base.json"))
	plan := PlanLibraryMoves(ws, projects, nil)
	if plan.Fix != nil || !strings.Contains(plan.Refusals["one"], "tsconfig.base.json") {
		t.Fatalf("%+v", plan)
	}
}

func TestPlanLibraryMovesIgnoresWhatIsNotALibraryWithExportsOrImporters(t *testing.T) {
	ws, projects := libraryWorkspace(t, map[string]string{
		"packages/one/package.json": `{"name":"@a/one","main":"./dist/index.js","types":"./dist/index.d.ts","scripts":{"build":"tsc --outDir dist"}}`,
	})
	plan := PlanLibraryMoves(ws, projects[:1], nil)
	if plan.Fix != nil || len(plan.Refusals) != 0 {
		t.Fatalf("a main-only package nobody imports takes the plain move: %+v", plan)
	}
	if plan = PlanLibraryMoves(ws, projects[:1], map[string][]string{"one": {"web"}}); plan.Fix == nil {
		t.Fatal("an imported package takes this move")
	}
}
