package manifestfixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func workspaceWith(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const rootManifest = `{"name": "w", "private": true, "workspaces": ["packages/*"], "devDependencies": {"eslint": "^10.0.0"}}`

func project(name string) protocol.Project {
	return protocol.Project{Name: name, Root: "packages/" + name}
}

func TestDeclareImportsFixTakesTheRangeTheWorkspaceUses(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":            rootManifest,
		"packages/a/package.json": "{\n  \"name\": \"a\",\n  \"dependencies\": {\n    \"axios\": \"^1.0.0\",\n    \"zod\": \"^3.0.0\"\n  }\n}\n",
		"packages/b/package.json": `{"name": "b", "dependencies": {"lodash": "^4.17.21"}}`,
		"packages/c/package.json": `{"name": "c", "dependencies": {"lodash": "^4.17.21"}}`,
	})
	fix := DeclareImportsFix(root, project("a"), []declaredimports.UndeclaredImport{
		{File: "src/index.ts", Package: "lodash"},
		{File: "src/index.spec.ts", Package: "eslint"},
		{File: "src/unknown.ts", Package: "nobody-declares-this"},
	})
	if fix == nil || fix.Safe || len(fix.Choices) != 0 || fix.Description != DeclareImportsDescription || len(fix.AppliesTo) != 1 || fix.AppliesTo[0] != "a" {
		t.Fatalf("%+v", fix)
	}
	if len(fix.Steps) != 1 || fix.Steps[0].Argv[1] != "install" {
		t.Fatalf("npm install follows the edit: %+v", fix.Steps)
	}
	got := fix.Files[0].Content
	want := "{\n  \"name\": \"a\",\n  \"dependencies\": {\n    \"axios\": \"^1.0.0\",\n    \"lodash\": \"^4.17.21\",\n    \"zod\": \"^3.0.0\"\n  },\n  \"devDependencies\": {\n    \"eslint\": \"^10.0.0\"\n  }\n}\n"
	if fix.Files[0].Path != "packages/a/package.json" || got != want {
		t.Fatalf("shipped code goes to dependencies, a spec to devDependencies, sorted; a package nobody declares is left out:\n%s", got)
	}
}

func TestDeclareImportsFixAsksWhichVersionWhenTheWorkspaceDisagrees(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":            rootManifest,
		"packages/a/package.json": `{"name": "a"}`,
		"packages/b/package.json": `{"name": "b", "dependencies": {"lodash": "^4.17.21"}}`,
		"packages/c/package.json": `{"name": "c", "dependencies": {"lodash": "^4.17.21"}}`,
		"packages/d/package.json": `{"name": "d", "dependencies": {"lodash": "^3.10.1"}}`,
	})
	fix := DeclareImportsFix(root, project("a"), []declaredimports.UndeclaredImport{{File: "src/index.ts", Package: "lodash"}})
	if fix == nil || len(fix.Choices) != 1 || len(fix.Files) != 0 || len(fix.Steps) != 1 {
		t.Fatalf("a question, and the install after it: %+v", fix)
	}
	options := fix.Choices[0].Options
	if len(options) != 2 || !strings.Contains(options[0].Label, "lodash@^4.17.21 (declared in 2 package.json files)") || !strings.Contains(options[1].Label, "^3.10.1") {
		t.Fatalf("the most used range is the first option: %+v", options)
	}
	if !strings.Contains(options[1].Fix.Files[0].Content, `"lodash": "^3.10.1"`) {
		t.Fatalf("each option writes its own range: %s", options[1].Fix.Files[0].Content)
	}
}

func TestDeclareImportsFixMovesAPackageOutOfDevDependenciesAndFallsBackToTheInstalledVersion(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":                    `{"name": "w", "workspaces": ["packages/*"]}`,
		"packages/a/package.json":         `{"name": "a", "devDependencies": {"chalk": "^5.0.0"}}`,
		"node_modules/kleur/package.json": `{"name": "kleur", "version": "4.1.5"}`,
	})
	fix := DeclareImportsFix(root, project("a"), []declaredimports.UndeclaredImport{
		{File: "src/index.ts", Package: "chalk", DevOnly: true},
		{File: "src/index.ts", Package: "kleur"},
	})
	if fix == nil || len(fix.Files) != 1 {
		t.Fatalf("%+v", fix)
	}
	got := fix.Files[0].Content
	if !strings.Contains(got, `"chalk": "^5.0.0"`) || !strings.Contains(got, `"kleur": "^4.1.5"`) || strings.Contains(got, "devDependencies") {
		t.Fatalf("chalk moves to dependencies, kleur takes the installed version:\n%s", got)
	}
	if DeclareImportsFix(root, project("a"), []declaredimports.UndeclaredImport{{File: "src/x.ts", Package: "ghost"}}) != nil {
		t.Fatal("a package with no known version has no fix")
	}
}

func TestFixesKeepTheLineEndingsOfThePackageJSON(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":            rootManifest,
		"packages/a/package.json": "{\r\n  \"name\": \"a\"\r\n}\r\n",
		"packages/b/package.json": `{"name": "b", "dependencies": {"lodash": "^4.17.21"}}`,
	})
	fix := DeclareImportsFix(root, project("a"), []declaredimports.UndeclaredImport{{File: "src/index.ts", Package: "lodash"}})
	got := fix.Files[0].Content
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") || !strings.Contains(got, "\r\n") {
		t.Fatalf("a CRLF file stays CRLF, so the diff shows only what changed: %q", got)
	}
}
