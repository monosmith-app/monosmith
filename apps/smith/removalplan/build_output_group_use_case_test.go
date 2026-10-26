package removalplan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

func TestEmbracingTheBuildMovesTscOutputsToTheWorkspaceDist(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"package.json":                `{"name":"root","workspaces":["packages/*"]}`,
		"packages/kit/package.json":   `{"name":"kit","main":"dist/index.js","scripts":{"build":"tsc -p tsconfig.json"}}`,
		"packages/kit/tsconfig.json":  `{"compilerOptions":{"outDir":"dist","strict":true}}`,
		"packages/site/package.json":  `{"name":"site","scripts":{"build":"vite build"}}`,
		"packages/site/tsconfig.json": `{"compilerOptions":{"outDir":"build"}}`,
	}
	var names []string
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(body), 0o644)
		names = append(names, name)
	}
	p := Build(legacyinventory.NewTree(dir, names), fullChoices, nil)
	g := group(t, p, GroupOutput)
	if g.Files != 0 || g.Packages != 0 {
		t.Fatalf("an output group deletes nothing: %+v", g)
	}
	edited := map[string]string{}
	for _, e := range g.Fix.Files {
		edited[e.Path] = e.Content
	}
	if len(edited) != 2 || strings.Contains(edited["packages/kit/tsconfig.json"], "outDir") ||
		!strings.Contains(edited["packages/kit/package.json"], "--outDir ../../dist/packages/kit") ||
		!strings.Contains(edited["packages/kit/package.json"], `"main": "../../dist/packages/kit/index.js"`) {
		t.Fatalf("kit moves, the vite site stays: %v", edited)
	}
	keep := fullChoices
	keep.Build = "keep"
	for _, g := range Build(legacyinventory.NewTree(dir, names), keep, nil).Groups {
		if g.ID == GroupOutput {
			t.Fatal("--build keep leaves the outputs where they are")
		}
	}
}
