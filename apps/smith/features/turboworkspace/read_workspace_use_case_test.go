package turboworkspace

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) (string, []string) {
	t.Helper()
	root := t.TempDir()
	var list []string
	for rel, text := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		list = append(list, rel)
	}
	sort.Strings(list)
	return root, list
}

func TestReadPnpmWorkspace(t *testing.T) {
	root, files := writeTree(t, map[string]string{
		"turbo.json": `{
  // comments are allowed
  "$schema": "https://turborepo.com/schema.json",
  "tasks": {"build": {"dependsOn": ["^build"], "outputs": ["dist/**"],},},
}`,
		"package.json":                  `{"name": "acme", "packageManager": "pnpm@9.1.0", "devDependencies": {"turbo": "^2.1.0"}}`,
		"pnpm-workspace.yaml":           "packages:\n  - 'apps/*'\n  - 'packages/*'\n  - '!packages/skipped'\ncatalog:\n  react: ^19.0.0\n",
		"apps/web/package.json":         `{"name": "web", "private": true}`,
		"packages/ui/package.json":      `{"name": "@acme/ui", "version": "1.2.0"}`,
		"packages/ui/turbo.json":        `{"extends": ["//"], "tasks": {"build": {"outputs": ["$TURBO_EXTENDS$", "css/**"]}}}`,
		"packages/skipped/package.json": `{"name": "skipped"}`,
		"tools/other/package.json":      `{"name": "outside"}`,
	})
	w, err := Read(root, files)
	if err != nil {
		t.Fatal(err)
	}
	if w.Manager != PNPM || w.TurboVersion != "^2.1.0" {
		t.Fatalf("manager %q, turbo %q", w.Manager, w.TurboVersion)
	}
	var dirs []string
	for _, p := range w.Packages {
		dirs = append(dirs, p.Dir)
	}
	if want := []string{"apps/web", "packages/ui"}; !reflect.DeepEqual(dirs, want) {
		t.Fatalf("packages %v, want %v", dirs, want)
	}
	if w.Pnpm == nil || w.Pnpm.Catalog["react"] != "^19.0.0" {
		t.Fatalf("pnpm-workspace.yaml: %+v", w.Pnpm)
	}
	if ui := w.PackageByName("@acme/ui"); ui == nil || ui.Turbo == nil {
		t.Fatal("the package configuration is read")
	}
	if got := w.Turbo.TaskSet()["build"].Outputs; !reflect.DeepEqual(got, []string{"dist/**"}) {
		t.Fatalf("turbo.json with comments and trailing commas: %v", got)
	}
	shape := w.NxShape()
	if len(shape.Projects) != 2 || shape.Projects[0].Type != "app" || shape.Projects[1].Type != "lib" {
		t.Fatalf("kinds: %+v", shape.Projects)
	}
}

func TestReadNeedsTurboJSON(t *testing.T) {
	root, files := writeTree(t, map[string]string{"package.json": `{}`})
	if _, err := Read(root, files); err != ErrNotTurbo {
		t.Fatalf("want ErrNotTurbo, got %v", err)
	}
}

func TestDetectManager(t *testing.T) {
	cases := []struct {
		field string
		files []string
		want  Manager
	}{
		{"yarn@4.1.0", nil, Yarn},
		{"", []string{"pnpm-lock.yaml"}, PNPM},
		{"", []string{"yarn.lock"}, Yarn},
		{"", []string{"package-lock.json"}, NPM},
		{"", []string{"pnpm-workspace.yaml"}, PNPM},
		{"", nil, NPM},
	}
	for _, c := range cases {
		if got := DetectManager(c.field, c.files); got != c.want {
			t.Errorf("%q %v: %s, want %s", c.field, c.files, got, c.want)
		}
	}
}
