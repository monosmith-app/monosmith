package legacyinventory

import (
	"strings"
	"testing"
)

// tree makes a Tree of in-memory files.
func tree(files map[string]string) *Tree {
	t := NewTree("", nil)
	for name, content := range files {
		t.Write(name, content)
	}
	return t
}

func paths(inv Inventory) string {
	var out []string
	for _, f := range inv.Files {
		out = append(out, f.Path)
	}
	return strings.Join(out, ",")
}

func TestScanFilesRemovesABuildConfigNothingRuns(t *testing.T) {
	tr := tree(map[string]string{
		"package.json":                 `{"scripts":{"build":"esbuild src/index.ts --bundle"}}`,
		"packages/a/rollup.config.cjs": "module.exports = {}",
		"packages/a/package.json":      `{"name":"a"}`,
	})
	inv := ScanFiles(tr, []Family{FamilyBuild})
	if got := paths(inv); got != "packages/a/rollup.config.cjs" {
		t.Fatalf("removable = %q, want the rollup config", got)
	}
}

func TestScanFilesKeepsABuildConfigAScriptRuns(t *testing.T) {
	tr := tree(map[string]string{
		"package.json":                 `{"scripts":{"build":"rollup -c"}}`,
		"packages/a/rollup.config.cjs": "module.exports = {}",
	})
	inv := ScanFiles(tr, []Family{FamilyBuild})
	if len(inv.Files) != 0 || len(inv.Kept) != 1 || !strings.Contains(inv.Kept[0].Why, "package.json") {
		t.Fatalf("want the config kept because package.json runs rollup, got %+v", inv)
	}
}

func TestScanFilesKeepsAConfigAnotherOneImports(t *testing.T) {
	tr := tree(map[string]string{
		"eslint.config.mjs":      "import base from './eslint.config.base.mjs'",
		"eslint.config.base.mjs": "export default []",
	})
	// Both are lint configs and both are candidates: the root one goes with
	// the lint family, so the base it imported has nothing left to keep it.
	inv := ScanFiles(tr, []Family{FamilyLint})
	if len(inv.Files) != 2 {
		t.Fatalf("want both lint configs removable, got %+v", inv)
	}
	// A file that stays and names the base keeps it.
	tr.Write("tools/check.mjs", "import './eslint.config.base.mjs'")
	inv = ScanFiles(tr, []Family{FamilyLint})
	if len(inv.Kept) != 1 || inv.Kept[0].What != "eslint.config.base.mjs" {
		t.Fatalf("want the base kept because tools/check.mjs imports it, got %+v", inv)
	}
}

func TestScanFilesKeepsTheRunnerStillInUse(t *testing.T) {
	tr := tree(map[string]string{
		"vitest.config.ts":           "export default {}",
		"src/a.spec.ts":              "import { it } from 'vitest'",
		"packages/b/vitest.setup.ts": "",
	})
	inv := ScanFiles(tr, []Family{FamilyVitest})
	if len(inv.Files) != 0 || len(inv.Kept) != 2 {
		t.Fatalf("a spec importing vitest keeps its configs, got %+v", inv)
	}
	tr.Delete("src/a.spec.ts")
	if inv := ScanFiles(tr, []Family{FamilyVitest}); len(inv.Files) != 2 {
		t.Fatalf("with no vitest spec left both configs go, got %+v", inv)
	}
}

func TestScanFilesKeepsSwcrcWhileJestLoadsIt(t *testing.T) {
	tr := tree(map[string]string{
		"packages/a/.swcrc":         "{}",
		"packages/a/jest.config.ts": "transform: { '^.+$': ['@swc/jest'] }",
	})
	if inv := ScanFiles(tr, []Family{FamilyBuild}); len(inv.Files) != 0 {
		t.Fatalf("@swc/jest reads .swcrc by default, got %+v", inv)
	}
	tr.Write("packages/a/jest.config.ts", "swcJestConfig.swcrc = false; transform: { '^.+$': ['@swc/jest'] }")
	if inv := ScanFiles(tr, []Family{FamilyBuild}); paths(inv) != "packages/a/.swcrc" {
		t.Fatalf("with swcrc turned off the .swcrc goes, got %+v", inv)
	}
}
