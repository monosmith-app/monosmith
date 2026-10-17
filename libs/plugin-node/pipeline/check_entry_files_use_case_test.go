package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const readerManifest = `{
  "name": "@acme/reader",
  "main": "./dist/index.esm.js",
  "types": "./dist/src/index.d.ts",
  "exports": {
    "./package.json": "./package.json",
    ".": {"types": "./dist/src/index.d.ts", "import": "./dist/index.esm.js"},
    "./xlsx": {"types": "./dist/src/spreadsheet/index.d.ts", "import": "./dist/xlsx.esm.js"},
    "./assets/*": "./dist/assets/*"
  }
}`

func readerWorkspace(t *testing.T, built ...string) (string, protocol.Project) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "packages", "reader")
	for _, f := range append([]string{"package.json"}, built...) {
		file := filepath.Join(dir, filepath.FromSlash(f))
		_ = os.MkdirAll(filepath.Dir(file), 0o755)
		content := ""
		if f == "package.json" {
			content = readerManifest
		}
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, protocol.Project{Name: "reader", Root: "packages/reader", Kind: protocol.ProjectPackage, Tasks: map[string]protocol.Task{"build": {}}}
}

func TestCheckEntryFilesNamesWhatTheBuildDidNotProduce(t *testing.T) {
	// OpenCraw's office-reader after the first conversion: only index built.
	root, p := readerWorkspace(t, "dist/index.esm.js", "dist/src/index.d.ts", "dist/src/spreadsheet/index.d.ts")
	f, ok := CheckEntryFiles(root, p)
	if !ok || f.OK || f.Detail != "packages/reader/package.json names dist/xlsx.esm.js, which the build did not produce" {
		t.Fatalf("finding %+v %v", f, ok)
	}
}

func TestCheckEntryFilesPassesACompleteBuild(t *testing.T) {
	root, p := readerWorkspace(t, "dist/index.esm.js", "dist/xlsx.esm.js", "dist/src/index.d.ts", "dist/src/spreadsheet/index.d.ts")
	if f, ok := CheckEntryFiles(root, p); !ok || !f.OK {
		t.Fatalf("finding %+v %v", f, ok)
	}
}

func TestCheckEntryFilesWaitsForABuild(t *testing.T) {
	root, p := readerWorkspace(t)
	if f, ok := CheckEntryFiles(root, p); ok {
		t.Fatalf("an unbuilt package is not judged: %+v", f)
	}
	p.Kind = protocol.ProjectApp
	root, _ = readerWorkspace(t, "dist/index.esm.js")
	if _, ok := CheckEntryFiles(root, p); ok {
		t.Fatal("only packages are installed by others")
	}
}
