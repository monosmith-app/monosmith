package convert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

func TestEditManifestsForAnotherSourceTool(t *testing.T) {
	root := t.TempDir()
	write := func(rel, text string) {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", `{
  "name": "acme",
  "packageManager": "pnpm@9.1.0",
  "engines": {"node": ">=20", "pnpm": ">=9"},
  "pnpm": {"overrides": {"semver": "7.5.2"}},
  "devDependencies": {"turbo": "^2.1.0", "@acme/ui": "workspace:*"}
}`)
	write("apps/web/package.json", `{"name": "web", "dependencies": {"@acme/ui": "workspace:*", "zod": "^3.0.0"}, "scripts": {"build": "next build"}}`)

	plan := &convertplan.Plan{
		Root:               root,
		RootScripts:        map[string]string{},
		RemoveFields:       []string{"packageManager", "engines.pnpm", "pnpm"},
		RemoveDependencies: []string{"turbo"},
		SetWorkspaces:      []string{"apps/*", "packages/*"},
		SetOverrides:       map[string]string{"semver": "7.5.2"},
	}
	plan.SetSpecifier(".", "@acme/ui", "*")
	plan.SetSpecifier("apps/web", "@acme/ui", "^1.2.0")
	pp := &convertplan.ProjectPlan{Name: "web", Root: "apps/web", Language: "node", Supported: true, Scripts: map[string]string{"build": "smith run build"}}

	if err := editRootPackage(root, plan, ""); err != nil {
		t.Fatal(err)
	}
	if err := editManifest(root, pp, plan.Specifiers); err != nil {
		t.Fatal(err)
	}

	rootText, _ := os.ReadFile(filepath.Join(root, "package.json"))
	for _, gone := range []string{"packageManager", `"pnpm"`, "turbo", "workspace:"} {
		if strings.Contains(string(rootText), gone) {
			t.Errorf("root package.json still has %q:\n%s", gone, rootText)
		}
	}
	for _, want := range []string{`"node": ">=20"`, `"workspaces"`, `"apps/*"`, `"overrides"`, `"semver": "7.5.2"`, `"@acme/ui": "*"`} {
		if !strings.Contains(string(rootText), want) {
			t.Errorf("root package.json lacks %q:\n%s", want, rootText)
		}
	}
	webText, _ := os.ReadFile(filepath.Join(root, "apps", "web", "package.json"))
	for _, want := range []string{`"@acme/ui": "^1.2.0"`, `"zod": "^3.0.0"`, `"build": "smith run build"`} {
		if !strings.Contains(string(webText), want) {
			t.Errorf("web package.json lacks %q:\n%s", want, webText)
		}
	}
}

func TestDeleteFieldDropsAnEmptiedParent(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"engines": {"pnpm": ">=9"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plan := &convertplan.Plan{Root: root, RootScripts: map[string]string{}, RemoveFields: []string{"engines.pnpm"}}
	if err := editRootPackage(root, plan, ""); err != nil {
		t.Fatal(err)
	}
	text, _ := os.ReadFile(filepath.Join(root, "package.json"))
	if strings.Contains(string(text), "engines") {
		t.Fatalf("an engines left empty goes too:\n%s", text)
	}
}

func TestConvertedRootDependsOnTheSmithThatConvertedIt(t *testing.T) {
	for version, want := range map[string]bool{"1.54.0": true, "dev": false, "1.54.0-rc.1": false} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name": "acme", "devDependencies": {"jest": "^29.0.0"}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := editRootPackage(root, &convertplan.Plan{Root: root, RootScripts: map[string]string{"build": "smith run build"}}, version); err != nil {
			t.Fatal(err)
		}
		text, _ := os.ReadFile(filepath.Join(root, "package.json"))
		if got := strings.Contains(string(text), `"monosmith": "`+version+`"`); got != want {
			t.Errorf("version %q: monosmith pinned = %v, want %v:\n%s", version, got, want, text)
		}
	}
}
