package removalplan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/legacyinventory"
)

func TestEmbracingTheBuildMovesAnAppsTscOutputAndLeavesALibraryInPlace(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"package.json":                `{"name":"root","workspaces":["packages/*","apps/*"]}`,
		"apps/svc/package.json":       `{"name":"svc","private":true,"main":"dist/index.js","scripts":{"build":"tsc -p tsconfig.json"}}`,
		"apps/svc/tsconfig.json":      `{"compilerOptions":{"outDir":"dist","strict":true}}`,
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
	if len(edited) != 2 || strings.Contains(edited["apps/svc/tsconfig.json"], "outDir") ||
		!strings.Contains(edited["apps/svc/package.json"], "--outDir ../../dist/apps/svc") ||
		!strings.Contains(edited["apps/svc/package.json"], `"main": "../../dist/apps/svc/index.js"`) {
		t.Fatalf("the app svc moves, the vite site stays: %v", edited)
	}
	if _, ok := edited["packages/kit/package.json"]; ok {
		t.Fatalf("the library kit keeps building in its own dist (ADR 0102): %v", edited)
	}
	keep := fullChoices
	keep.Build = "keep"
	for _, g := range Build(legacyinventory.NewTree(dir, names), keep, nil).Groups {
		if g.ID == GroupOutput {
			t.Fatal("--build keep leaves the outputs where they are")
		}
	}
}

func TestEmbracingTheBuildLeavesAnImportedAppUnmovedAndSaysWhy(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"package.json":            `{"name":"root","workspaces":["apps/*"]}`,
		"apps/util/package.json":  `{"name":"@d/util","private":true,"main":"dist/index.js","types":"dist/index.d.ts","scripts":{"build":"tsc -p tsconfig.json"}}`,
		"apps/util/tsconfig.json": `{"compilerOptions":{"outDir":"dist"}}`,
		"apps/web/package.json":   `{"name":"web","private":true,"dependencies":{"@d/util":"*"},"scripts":{"build":"tsc -p tsconfig.json"}}`,
		"apps/web/tsconfig.json":  `{"compilerOptions":{"outDir":"dist"}}`,
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
	if edited["apps/util/package.json"] || edited["apps/util/tsconfig.json"] || !edited["apps/web/package.json"] {
		t.Fatalf("web moves, the imported util stays: %v", edited)
	}
	notes := strings.Join(p.Notes, "\n")
	if !strings.Contains(notes, "@d/util (apps/util) stays unmoved") || !strings.Contains(notes, "web import it by name") {
		t.Fatalf("the reason is a note of the review:\n%s", notes)
	}
}

func TestEmbracingTheBuildMovesTheDeclaredBuildOutputsToo(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"package.json":               `{"name":"root","workspaces":["apps/*"]}`,
		"apps/api/package.json":      `{"name":"api","private":true,"main":"./dist/functions.js","scripts":{"build":"tsc -p tsconfig.app.json && npm run link:prod"}}`,
		"apps/api/tsconfig.app.json": `{"compilerOptions":{"rootDir":"./src","outDir":"./dist"}}`,
		"smith.toml":                 "schema = 1\n\n[workspace]\nname = \"acme\"\n\n[projects.api.tasks.build]\ncommand = [\"npm\", \"run\", \"build\", \"-w\", \"api\"]\ncwd = \".\"\noutputs = [\"dist\"]\n\n[projects.api.tasks.build-dev]\ncommand = [\"npm\", \"run\", \"build:dev\", \"-w\", \"api\"]\noutputs = [\"dist\"]\n",
	}
	var names []string
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(body), 0o644)
		names = append(names, name)
	}
	g := group(t, Build(legacyinventory.NewTree(dir, names), fullChoices, nil), GroupOutput)
	config := ""
	for _, e := range g.Fix.Files {
		if e.Path == "smith.toml" {
			config = e.Content
		}
	}
	if !strings.Contains(config, "cwd = \".\"\noutputs = [\"../../dist/apps/api\"]") {
		t.Fatalf("the build's outputs follow the moved folder:\n%s", config)
	}
	if !strings.Contains(config, "\"build:dev\", \"-w\", \"api\"]\noutputs = [\"dist\"]") {
		t.Fatalf("a task that runs another script keeps its outputs:\n%s", config)
	}
	if !strings.Contains(g.Reason, "and so do the build outputs in smith.toml") {
		t.Fatalf("the reason names smith.toml: %s", g.Reason)
	}
}
