package corefixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddLintTaskFixDeclaresOneTaskPerProjectInSmithToml(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("smith.toml", "[workspace]\nname = \"w\"\n\n[projects.ms.sharepoint.tasks.build]\noutputs = [\"dist\"]\n")
	if AddLintTaskFix(root, []string{"web"}) != nil {
		t.Fatal("without an ESLint config at the root smith cannot say how to lint")
	}
	write("eslint.config.mjs", "export default []\n")
	fix := AddLintTaskFix(root, []string{"web", "ms.sharepoint"})
	if fix == nil || fix.Safe || len(fix.Files) != 1 || len(fix.AppliesTo) != 2 {
		t.Fatalf("%+v", fix)
	}
	text := fix.Files[0].Content
	for _, want := range []string{
		"[projects.web.tasks.lint]", `command = ["npx", "eslint", "."]`, `inputs = ["**/*", "{workspaceRoot}/eslint.config.mjs"]`,
		`[projects."ms.sharepoint".tasks.lint]`, "[projects.ms.sharepoint.tasks.build]",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("smith.toml lacks %q:\n%s", want, text)
		}
	}
	if AddLintTaskFix(root, nil) != nil {
		t.Fatal("no projects, no fix")
	}
}
