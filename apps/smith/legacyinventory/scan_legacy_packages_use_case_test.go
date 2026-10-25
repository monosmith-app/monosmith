package legacyinventory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func names(inv Inventory) string {
	var out []string
	for _, p := range inv.Packages {
		out = append(out, p.Name)
	}
	return strings.Join(out, ",")
}

const lintManifest = `{
  "devDependencies": {
    "eslint": "^10",
    "@mnci/eslint-config": "*",
    "typescript-eslint": "^8",
    "eslint-plugin-jest": "^29",
    "rollup": "^4",
    "left-pad": "^1"
  }
}`

func TestScanPackagesRemovesOnlyTheRetiredFamilyNothingUses(t *testing.T) {
	tr := tree(map[string]string{"package.json": lintManifest, "eslint.config.mjs": "import x from 'eslint-plugin-jest'"})
	inv := ScanPackages(tr, []Family{FamilyLint})
	if got := names(inv); got != "@mnci/eslint-config,typescript-eslint" {
		t.Fatalf("removable = %q: eslint itself, other families and unknown packages stay, and the jest plugin is imported", got)
	}
	if len(inv.Kept) != 1 || inv.Kept[0].What != "eslint-plugin-jest" {
		t.Fatalf("want eslint-plugin-jest kept, got %+v", inv.Kept)
	}
}

func TestScanPackagesIgnoresTheDependencySections(t *testing.T) {
	tr := tree(map[string]string{"package.json": lintManifest})
	if got := names(ScanPackages(tr, []Family{FamilyLint})); !strings.Contains(got, "eslint-plugin-jest") {
		t.Fatalf("a package only declared is not used, got %q", got)
	}
}

func TestScanPackagesKeepsAPackageAScriptRuns(t *testing.T) {
	tr := tree(map[string]string{"package.json": `{"scripts":{"build":"rollup -c"},"devDependencies":{"rollup":"^4"}}`})
	inv := ScanPackages(tr, []Family{FamilyBuild})
	if len(inv.Packages) != 0 || len(inv.Kept) != 1 {
		t.Fatalf("rollup is run by a script, got %+v", inv)
	}
}

func TestScanPackagesNeverTouchesAPeerDependency(t *testing.T) {
	tr := tree(map[string]string{
		"package.json":     `{"devDependencies":{"typescript-eslint":"^8"}}`,
		"lib/package.json": `{"peerDependencies":{"typescript-eslint":"^8"}}`,
	})
	inv := ScanPackages(tr, []Family{FamilyLint})
	if len(inv.Packages) != 0 || len(inv.Kept) != 1 || !strings.Contains(inv.Kept[0].Why, "peer dependency") {
		t.Fatalf("a peer dependency is a published contract, got %+v", inv)
	}
}

func TestScanPackagesKeepsARequiredPeerOfAPackageThatStays(t *testing.T) {
	dir := t.TempDir()
	writeInstalled(t, dir, "ts-loader-user", `{"name":"ts-loader-user","peerDependencies":{"webpack":"^5"}}`)
	tr := tree(map[string]string{"package.json": `{"devDependencies":{"ts-loader-user":"1","webpack":"^5"}}`})
	tr.Modules = dir
	inv := ScanPackages(tr, []Family{FamilyBuild})
	if len(inv.Packages) != 0 || len(inv.Kept) != 1 || !strings.Contains(inv.Kept[0].Why, "required peer of ts-loader-user") {
		t.Fatalf("webpack is a required peer of a package that stays, got %+v", inv)
	}
}

func TestScanPackagesKeepsTypesWhileTheirPackageIsUsed(t *testing.T) {
	tr := tree(map[string]string{
		"package.json":  `{"devDependencies":{"@types/jest":"^30"}}`,
		"src/a.test.ts": "jest.fn()",
	})
	// Jest is the retired family, but a test still uses its globals.
	if inv := ScanPackages(tr, []Family{FamilyJest}); len(inv.Packages) != 0 {
		t.Fatalf("a test uses jest, so @types/jest stays, got %+v", inv)
	}
}

func TestScanPackagesRemovesNxEverywhere(t *testing.T) {
	tr := tree(map[string]string{
		"package.json":            `{"devDependencies":{"nx":"1","@nx/js":"1"}}`,
		"packages/a/package.json": `{"devDependencies":{"@nx/eslint":"1"}}`,
	})
	inv := ScanPackages(tr, []Family{FamilyNx})
	if got := names(inv); got != "@nx/eslint,@nx/js,nx" {
		t.Fatalf("removable = %q", got)
	}
}

func writeInstalled(t *testing.T, dir, name, manifest string) {
	t.Helper()
	path := filepath.Join(dir, "node_modules", filepath.FromSlash(name))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}
