package removalplan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

func TestEmbracingTheBuildLeavesAnImportedPackageUnmovedAndSaysWhy(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"package.json":                `{"name":"root","workspaces":["packages/*"]}`,
		"packages/util/package.json":  `{"name":"@d/util","private":true,"main":"dist/index.js","types":"dist/index.d.ts","scripts":{"build":"tsc -p tsconfig.json"}}`,
		"packages/util/tsconfig.json": `{"compilerOptions":{"outDir":"dist"}}`,
		"packages/web/package.json":   `{"name":"web","private":true,"dependencies":{"@d/util":"*"},"scripts":{"build":"tsc -p tsconfig.json"}}`,
		"packages/web/tsconfig.json":  `{"compilerOptions":{"outDir":"dist"}}`,
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
	edited := map[string]bool{}
	for _, e := range g.Fix.Files {
		edited[e.Path] = true
	}
	if edited["packages/util/package.json"] || edited["packages/util/tsconfig.json"] || !edited["packages/web/package.json"] {
		t.Fatalf("web moves, the imported util stays: %v", edited)
	}
	notes := strings.Join(p.Notes, "\n")
	if !strings.Contains(notes, "@d/util (packages/util) stays unmoved") || !strings.Contains(notes, "web import it by name") {
		t.Fatalf("the reason is a note of the review:\n%s", notes)
	}
}
