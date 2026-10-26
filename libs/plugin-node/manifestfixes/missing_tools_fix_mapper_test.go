package manifestfixes

import (
	"strings"
	"testing"
)

func TestMissingToolsFixAsksToInstallOrRemoveTheScript(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json": "{\n  \"name\": \"w\",\n  \"scripts\": {\n    \"build\": \"tsc\",\n    \"format:check\": \"oxfmt --check .\"\n  }\n}\n",
	})
	fix := MissingToolsFix(root, "", "the root package.json", map[string][]string{"oxfmt": {"format:check"}})
	if fix == nil || fix.Safe || len(fix.Choices) != 1 {
		t.Fatalf("%+v", fix)
	}
	options := fix.Choices[0].Options
	if len(options) != 2 || options[0].Label != "install oxfmt: npm i -D oxfmt" || !strings.Contains(options[1].Label, "format:check (oxfmt --check .)") {
		t.Fatalf("%+v", options)
	}
	if got := strings.Join(options[0].Fix.Steps[0].Argv, " "); got != "npm install -D --no-audit --no-fund oxfmt" {
		t.Fatalf("%s", got)
	}
	removed := options[1].Fix.Files[0].Content
	if options[1].Fix.Files[0].Path != "package.json" || strings.Contains(removed, "oxfmt") || !strings.Contains(removed, `"build": "tsc"`) {
		t.Fatalf("only the script that starts the tool goes:\n%s", removed)
	}
}

func TestMissingToolsFixInstallsIntoTheProjectsWorkspace(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":            rootManifest,
		"packages/a/package.json": `{"name": "a", "scripts": {"fmt": "prettier ."}}`,
	})
	fix := MissingToolsFix(root, "packages/a", "a", map[string][]string{"prettier": {"fmt"}})
	got := strings.Join(fix.Choices[0].Options[0].Fix.Steps[0].Argv, " ")
	if got != "npm install -D --no-audit --no-fund --workspace packages/a prettier" {
		t.Fatalf("%s", got)
	}
	if MissingToolsFix(root, "", "root", nil) != nil {
		t.Fatal("no tools, no fix")
	}
}
