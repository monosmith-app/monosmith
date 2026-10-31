//go:build e2e

package goldenpath

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// siblingPackage is a package as a converted Nx workspace has it: esbuild
// builds it into dist, @swc/jest tests it, and three tsconfig files describe it.
func siblingPackage(name string) map[string]string {
	dist := "../../dist/packages/" + name
	return map[string]string{
		"packages/" + name + "/package.json": fmt.Sprintf(`{
  "name": "@nxlayout/%[1]s",
  "version": "0.0.0",
  "type": "module",
  "smith": { "kind": "package" },
  "main": "%[2]s/index.js",
  "types": "%[2]s/index.d.ts",
  "publishConfig": { "access": "public" },
  "scripts": {
    "build": "esbuild src/index.ts --bundle --packages=external --platform=node --format=esm --outdir=%[2]s && tsc -p tsconfig.build.json --declaration --emitDeclarationOnly --outDir %[2]s",
    "test": "jest",
    "typecheck": "tsc -p tsconfig.json --noEmit"
  },
  "jest": { "transform": { "^.+\\.ts$": ["@swc/jest", {}] } }
}
`, name, dist),
		"packages/" + name + "/src/index.ts":        "export const " + name + " = 1\n",
		"packages/" + name + "/tsconfig.json":       "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"compilerOptions\": { \"rootDir\": \"src\", \"types\": [\"node\"] },\n  \"include\": [\"src\"]\n}\n",
		"packages/" + name + "/tsconfig.build.json": "{ \"extends\": \"./tsconfig.json\" }\n",
		"packages/" + name + "/tsconfig.spec.json":  "{ \"extends\": \"./tsconfig.json\" }\n",
	}
}

// TestAddFollowsTheDialectOfSiblings: in a workspace laid out like a converted
// Nx one, whose packages build with esbuild into dist and test with @swc/jest,
// smith add package gives the new package the same build, the same test runner
// and the same entry points, says why, asks nothing, and warns about the one
// thing its stack cannot follow (the third tsconfig file). --fresh gives the
// stack's own default.
func TestAddFollowsTheDialectOfSiblings(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "nxlayout", "--ci", "github", "--default-stack", "node=node-jest")
	w.dir = filepath.Join(base, "nxlayout")
	w.write("package.json", `{
  "name": "nxlayout-workspace",
  "private": true,
  "workspaces": ["packages/*", "apps/*"]
}
`)
	for _, name := range []string{"alpha", "beta", "gamma"} {
		for rel, body := range siblingPackage(name) {
			w.write(rel, body)
		}
	}

	out := w.smith(true, "add", "package", "delta", "--lang", "node")
	for _, want := range []string{
		"The 3 sibling packages agree on:",
		"test.runner = @swc/jest (3 of 3 packages)",
		"build.tool = esbuild (3 of 3 packages)",
		"exports.target = dist (3 of 3 packages)",
		"Warning: 3 of 3 sibling packages have tsconfig.files = tsconfig.build.json,tsconfig.json,tsconfig.spec.json",
		"Done: package delta is ready",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output lacks %q:\n%s", want, out)
		}
	}
	manifest := w.read("packages/delta/package.json")
	for _, want := range []string{`"build": "esbuild src/index.ts`, `"main": "../../dist/packages/delta/index.js"`, `"@swc/jest"`} {
		if !strings.Contains(manifest, want) {
			t.Errorf("packages/delta/package.json lacks %s:\n%s", want, manifest)
		}
	}
	if strings.Contains(manifest, "ts-jest") {
		t.Errorf("the new package tests with ts-jest:\n%s", manifest)
	}
	if doctor := w.smith(true, "doctor"); strings.Contains(doctor, "delta") && strings.Contains(doctor, "✖") {
		t.Errorf("doctor finds a problem with the new package:\n%s", doctor)
	}

	fresh := w.smith(true, "add", "package", "epsilon", "--lang", "node", "--fresh")
	if strings.Contains(fresh, "sibling") {
		t.Errorf("--fresh looks at the siblings:\n%s", fresh)
	}
	if manifest := w.read("packages/epsilon/package.json"); !strings.Contains(manifest, `"build": "tsc -p tsconfig.build.json`) || strings.Contains(manifest, "esbuild") {
		t.Errorf("--fresh does not build with tsc:\n%s", manifest)
	}
}
