package adoption

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestTheProgramsAndLibrariesOfARootModuleAreDeclared(t *testing.T) {
	root := writeTree(t, map[string]string{
		"go.mod":                     "module example.com/mvd\n\ngo 1.22\n",
		"apps/mvd/main.go":           "package main\n",
		"cmd/mvd-tui/main.go":        "package main\n",
		"libs/core/core.go":          "package core\n",
		"libs/core/greeting/g.go":    "package greeting\n",
		"pkg/util/u.go":              "package util\n",
		"internal/store/s.go":        "package store\n",
		"libs/tools/cmd/gen/main.go": "package main\n",
		"vendor/x/x.go":              "package x\n",
		"testdata/t/main.go":         "package main\n",
		"tools/other/go.mod":         "module example.com/other\n",
		"tools/other/main.go":        "package main\n",
		"apps/mvd/main_test.go":      "package main\n",
		"docs/example.go":            "package main\n",
	})
	got, rootProgram := InferGoProjects(root)
	var summary []string
	for _, p := range got {
		summary = append(summary, p.Name+" "+p.Kind+" "+p.Root)
	}
	want := []string{
		"mvd app apps/mvd",
		"mvd-tui app cmd/mvd-tui",
		"docs app docs",
		"store lib internal/store",
		"core lib libs/core",
		"gen app libs/tools/cmd/gen",
		"util lib pkg/util",
	}
	if rootProgram || !reflect.DeepEqual(summary, want) {
		t.Fatalf("rootProgram %v\n got %q\nwant %q", rootProgram, summary, want)
	}
}

func TestARepositoryWithoutARootModuleOrWithAProgramAtTheRootIsReported(t *testing.T) {
	if got, _ := InferGoProjects(writeTree(t, map[string]string{"apps/a/main.go": "package main\n"})); got != nil {
		t.Errorf("no root go.mod, nothing to declare: %v", got)
	}
	root := writeTree(t, map[string]string{"go.mod": "module x\n", "main.go": "package main\n", "libs/a/a.go": "package a\n"})
	got, rootProgram := InferGoProjects(root)
	if !rootProgram || len(got) != 1 || got[0].Root != "libs/a" {
		t.Errorf("a program in the module's own folder cannot be declared: %v %v", got, rootProgram)
	}
}

func TestTwoFoldersWithOneNameGetTheirWholePath(t *testing.T) {
	root := writeTree(t, map[string]string{"go.mod": "module x\n", "apps/web/main.go": "package main\n", "cmd/web/main.go": "package main\n"})
	got, _ := InferGoProjects(root)
	if len(got) != 2 || got[0].Name != "web" || got[1].Name != "cmd-web" {
		t.Errorf("%+v", got)
	}
	text := declarationsText(got)
	if !strings.Contains(text, "[projects.web]\nroot = \"apps/web\"\nlanguage = \"go\"\nkind = \"app\"\n") {
		t.Errorf("%q", text)
	}
}
