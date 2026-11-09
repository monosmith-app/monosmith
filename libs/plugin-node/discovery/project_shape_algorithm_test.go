package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func shapeOf(t *testing.T, files map[string]string) map[string]string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	raw := []byte(files["package.json"])
	var pkg PackageManifest
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	return ProjectShape(dir, raw, pkg)
}

func TestProjectShapeOfAnSwcEsbuildLibraryWithOwnedConfigs(t *testing.T) {
	shape := shapeOf(t, map[string]string{
		"package.json": `{"name":"@acme/a","scripts":{"build":"esbuild src/index.ts --bundle","test":"jest"},
			"main":"./dist/index.js","jest":{"transform":{"^.+\\.ts$":["@swc/jest",{}]}}}`,
		"tsconfig.json":      "{}",
		"tsconfig.lib.json":  "{}",
		"eslint.config.mjs":  "",
		"src/index.ts":       "",
		"node_modules/x/y.j": "",
	})
	want := map[string]string{
		ShapeBuildTool: "esbuild", ShapeTestRunner: "@swc/jest", ShapeExportsTarget: "dist",
		ShapeTsconfigFiles: "tsconfig.json,tsconfig.lib.json", ShapeLintConfig: "project",
		ShapeScripts: "build,test", ShapeLayout: "src",
	}
	for key, value := range want {
		if shape[key] != value {
			t.Errorf("%s = %q, want %q", key, shape[key], value)
		}
	}
}

func TestProjectShapeOfABarePackage(t *testing.T) {
	shape := shapeOf(t, map[string]string{"package.json": `{"name":"bare"}`})
	for _, key := range []string{ShapeBuildTool, ShapeTestRunner, ShapeExportsTarget, ShapeTsconfigFiles, ShapeScripts, ShapeLayout} {
		if shape[key] != ShapeNone {
			t.Errorf("%s = %q, want none", key, shape[key])
		}
	}
	if shape[ShapeLintConfig] != "root" {
		t.Errorf("lint.config = %q, want root", shape[ShapeLintConfig])
	}
}

func TestProjectShapeTellsTheTestRunnersApart(t *testing.T) {
	cases := map[string]struct {
		files map[string]string
		want  string
	}{
		"ts-jest in a config file": {map[string]string{"package.json": `{"scripts":{"test":"jest"}}`, "jest.config.js": "preset: 'ts-jest'"}, "ts-jest"},
		"plain jest":               {map[string]string{"package.json": `{"scripts":{"test":"npx jest --ci"}}`}, "jest"},
		"vitest":                   {map[string]string{"package.json": `{"scripts":{"test":"vitest run"}}`}, "vitest"},
		"node test":                {map[string]string{"package.json": `{"scripts":{"test":"node --test"}}`}, "node:test"},
		"another program":          {map[string]string{"package.json": `{"scripts":{"test":"mocha"}}`}, "mocha"},
	}
	for name, c := range cases {
		if got := shapeOf(t, c.files)[ShapeTestRunner]; got != c.want {
			t.Errorf("%s: test.runner = %q, want %q", name, got, c.want)
		}
	}
}

func TestProjectShapeFindsWhereTheEntryPointsLead(t *testing.T) {
	cases := map[string]struct{ manifest, want string }{
		"exports string in src":      {`{"exports":"./src/index.ts"}`, "src"},
		"exports condition in dist":  {`{"exports":{".":{"types":"./src/x.d.ts","import":"./dist/index.mjs"}}}`, "dist"},
		"main in the workspace dist": {`{"main":"../../dist/libs/a/index.js"}`, "dist"},
		"somewhere else":             {`{"main":"lib/index.js"}`, "other"},
		"no entry point":             {`{}`, "none"},
	}
	for name, c := range cases {
		if got := shapeOf(t, map[string]string{"package.json": c.manifest})[ShapeExportsTarget]; got != c.want {
			t.Errorf("%s: exports.target = %q, want %q", name, got, c.want)
		}
	}
}

func TestProjectShapeSaysWhereTheBuildWritesAndHowJestIsConfigured(t *testing.T) {
	for _, c := range []struct {
		name   string
		files  map[string]string
		output string
		mapped string
		jest   string
	}{
		{"own dist, exports map, config file", map[string]string{
			"package.json":    `{"scripts":{"build":"esbuild src/index.ts --outfile=dist/index.esm.js && tsc -p tsconfig.lib.json --declarationDir dist/src"},"exports":{".":"./dist/index.esm.js"}}`,
			"jest.config.cts": "",
		}, ShapeOutputProject, ShapeMapped, ShapeConfigFile},
		{"workspace dist, inline jest", map[string]string{
			"package.json": `{"scripts":{"build":"esbuild src/index.ts --outdir=../../dist/packages/a && tsc --outDir ../../dist/packages/a"},"exports":"./src/index.ts","jest":{}}`,
		}, ShapeOutputWorkspace, ShapeNone, ShapeConfigPackage},
		{"nothing", map[string]string{"package.json": `{"scripts":{"build":"vite build"}}`}, ShapeNone, ShapeNone, ShapeNone},
	} {
		shape := shapeOf(t, c.files)
		if shape[ShapeBuildOutput] != c.output || shape[ShapeExportsMap] != c.mapped || shape[ShapeJestConfig] != c.jest {
			t.Errorf("%s: got %s %s %s, want %s %s %s", c.name, shape[ShapeBuildOutput], shape[ShapeExportsMap], shape[ShapeJestConfig], c.output, c.mapped, c.jest)
		}
	}
}
