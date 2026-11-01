//go:build e2e

package goldenpath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// nxPackage is a package as a converted Nx workspace has it: tsconfig.json
// only references tsconfig.lib.json and tsconfig.spec.json, esbuild builds the
// code, @swc/jest reads jest.config.cts and .spec.swcrc, and out is where the
// build writes, relative to the package.
func nxPackage(name, out string) map[string]string {
	base := "packages/" + name + "/"
	return map[string]string{
		base + "package.json": fmt.Sprintf(`{
  "name": "@nxlayout/%[1]s",
  "version": "0.0.0",
  "type": "module",
  "smith": { "kind": "package" },
  "main": "%[2]s/index.js",
  "types": "%[2]s/index.d.ts",
  "publishConfig": { "access": "public" },
  "scripts": {
    "build": "esbuild src/index.ts --bundle --packages=external --platform=node --format=esm --outdir=%[2]s && tsc -p tsconfig.lib.json --declaration --emitDeclarationOnly --declarationDir %[2]s",
    "lint": "eslint .",
    "test": "jest",
    "typecheck": "tsc -p tsconfig.lib.json --noEmit"
  }
}
`, name, out),
		base + "src/index.ts": "export const " + name + " = 1\n",
		base + "tsconfig.json": `{
  "extends": "../../tsconfig.base.json",
  "files": [],
  "include": [],
  "references": [{ "path": "./tsconfig.lib.json" }, { "path": "./tsconfig.spec.json" }]
}
`,
		base + "tsconfig.lib.json":  "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"compilerOptions\": { \"rootDir\": \"src\", \"types\": [\"node\"] },\n  \"include\": [\"src\"],\n  \"exclude\": [\"src/**/*.spec.ts\"]\n}\n",
		base + "tsconfig.spec.json": "{\n  \"extends\": \"../../tsconfig.base.json\",\n  \"compilerOptions\": { \"rootDir\": \"src\", \"noEmit\": true, \"types\": [\"node\"] },\n  \"include\": [\"src/**/*.spec.ts\"]\n}\n",
		base + ".spec.swcrc":        "{ \"jsc\": { \"parser\": { \"syntax\": \"typescript\" }, \"target\": \"es2022\" }, \"module\": { \"type\": \"commonjs\" } }\n",
		base + "jest.config.cts":    "const { readFileSync } = require('fs')\nconst swcJestConfig = JSON.parse(readFileSync(`${__dirname}/.spec.swcrc`, 'utf-8'))\nswcJestConfig.swcrc = false\nmodule.exports = { testEnvironment: 'node', transform: { '^.+\\\\.ts$': ['@swc/jest', swcJestConfig] } }\n",
	}
}

// nxWorkspace is a converted Nx workspace of three packages whose builds write
// to out(name), and its root files.
func nxWorkspace(t *testing.T, out func(name string) string, smithToml string) *workspace {
	t.Helper()
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
	if smithToml != "" {
		f, err := os.OpenFile(filepath.Join(w.dir, "smith.toml"), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString(smithToml); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}
	for _, name := range []string{"alpha", "beta", "gamma"} {
		for rel, body := range nxPackage(name, out(name)) {
			w.write(rel, body)
		}
	}
	return w
}

// assertFollowsNxSiblings checks the package bar has what its siblings have, that
// smith add warned about nothing, and that smith doctor finds nothing wrong.
func assertFollowsNxSiblings(t *testing.T, w *workspace, added string) {
	t.Helper()
	for _, want := range []string{
		"The 3 sibling packages agree on:",
		"tsconfig.files = tsconfig.json,tsconfig.lib.json,tsconfig.spec.json (3 of 3 packages)",
		"Done: package bar is ready",
	} {
		if !strings.Contains(added, want) {
			t.Errorf("the output lacks %q:\n%s", want, added)
		}
	}
	if strings.Contains(added, "Warning:") {
		t.Errorf("smith add warns although the package follows its siblings:\n%s", added)
	}
	for _, rel := range []string{"tsconfig.json", "tsconfig.lib.json", "tsconfig.spec.json", ".spec.swcrc", "jest.config.cts"} {
		if _, err := os.Stat(filepath.Join(w.dir, "packages", "bar", rel)); err != nil {
			t.Errorf("packages/bar/%s is missing", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "bar", "tsconfig.build.json")); err == nil {
		t.Error("packages/bar has a tsconfig.build.json that no sibling has")
	}
	if manifest := w.read("packages/bar/package.json"); strings.Contains(manifest, `"jest": {`) {
		t.Errorf("the Jest configuration is inline although the siblings keep it in files:\n%s", manifest)
	}
	if doctor := w.smith(true, "doctor"); strings.Contains(doctor, "✖") {
		t.Errorf("doctor finds a problem after smith add:\n%s", doctor)
	}
}

// TestAddFollowsTheTsconfigSetAndTheBuildOutputOfAnNxLayout: in a converted Nx
// workspace whose packages write under dist/packages/<name>, smith add package
// gives the new package the same tsconfig set, the same Jest files and the same
// build output, warns about nothing, and leaves smith doctor green. It also
// says why a lib goes to packages/.
func TestAddFollowsTheTsconfigSetAndTheBuildOutputOfAnNxLayout(t *testing.T) {
	w := nxWorkspace(t, func(name string) string { return "../../dist/packages/" + name }, "")
	added := w.smith(true, "add", "package", "bar", "--lang", "node")
	assertFollowsNxSiblings(t, w, added)
	if manifest := w.read("packages/bar/package.json"); !strings.Contains(manifest, "--outdir=../../dist/packages/bar") {
		t.Errorf("the build does not write under dist/packages/bar:\n%s", manifest)
	}
	lib := w.smith(true, "add", "lib", "baz", "--lang", "node")
	if !strings.Contains(lib, "The lib goes to packages/, not libs/") {
		t.Errorf("smith add lib does not say why it chose packages/:\n%s", lib)
	}
}

// TestAddFollowsSiblingsThatBuildIntoTheirOwnDist: the build of every sibling
// writes to packages/<name>/dist and the workspace declares dist as the build
// outputs of its Node projects. The new package builds there too, so the task's
// cache and the build agree and smith doctor stays green.
func TestAddFollowsSiblingsThatBuildIntoTheirOwnDist(t *testing.T) {
	w := nxWorkspace(t, func(string) string { return "dist" }, "\n[[task_groups]]\nselect = [\"lang:node\"]\n\n[task_groups.tasks.build]\noutputs = [\"dist\"]\n")
	added := w.smith(true, "add", "package", "bar", "--lang", "node")
	assertFollowsNxSiblings(t, w, added)
	manifest := w.read("packages/bar/package.json")
	if !strings.Contains(manifest, "--outdir=dist") || strings.Contains(manifest, "../../dist") {
		t.Errorf("the build does not write into the package's own dist:\n%s", manifest)
	}
}
