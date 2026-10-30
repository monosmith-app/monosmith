package declaredimports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lines(r Result) string {
	var out []string
	for _, i := range r.Issues {
		out = append(out, i.String())
	}
	return strings.Join(out, "\n")
}

const goMod = `module example.com/app

go 1.22

require (
	example.com/direct v1.0.0
	example.com/unused v1.2.0
	example.com/marked v0.1.0 // indirect
	example.com/other/v2 v2.0.0
)
`

func TestCompareFindsMissingUnusedAndMarkedIndirect(t *testing.T) {
	root := t.TempDir()
	write(t, root, "app/go.mod", goMod)
	write(t, root, "app/main.go", `package main

import (
	"fmt"

	"example.com/direct/sub"
	"example.com/marked/x"
	"example.com/other/v2/api"
	"example.com/app/inner"
	"example.com/nowhere/pkg"
)
`)
	write(t, root, "app/inner/inner.go", "package inner\n")

	result, leans, err := Compare(root, "app")
	if err != nil {
		t.Fatal(err)
	}
	if leans {
		t.Error("no go.work, so the module leans on nothing")
	}
	got := lines(result)
	for _, want := range []string{
		"main.go imports example.com/nowhere/pkg, which no requirement in go.mod provides",
		"main.go imports example.com/marked, which go.mod marks as indirect",
		"go.mod requires example.com/unused, but nothing imports it",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(result.Issues) != 3 {
		t.Errorf("want exactly 3 issues, got:\n%s", got)
	}
}

func TestCompareCountsTestsToolsGenerateAndLeavesOutIgnoredAndNested(t *testing.T) {
	root := t.TempDir()
	write(t, root, "app/go.mod", `module example.com/app

go 1.24

require (
	example.com/test-only v1.0.0
	example.com/tool v1.0.0
	example.com/generator v1.0.0
	example.com/skipped v1.0.0
)

tool example.com/tool/cmd/t
`)
	write(t, root, "app/a_test.go", "package app\n\nimport _ \"example.com/test-only\"\n")
	write(t, root, "app/gen.go", "package app\n\n//go:generate go run example.com/generator/cmd/g@latest -x\n")
	write(t, root, "app/scratch.go", "//go:build ignore\n\npackage main\n\nimport _ \"example.com/ignored-tag\"\n")
	write(t, root, "app/testdata/x.go", "package x\n\nimport _ \"example.com/skipped\"\n")
	write(t, root, "app/nested/go.mod", "module example.com/app/nested\n")
	write(t, root, "app/nested/n.go", "package nested\n\nimport _ \"example.com/nested-dep\"\n")

	result, _, err := Compare(root, "app")
	if err != nil {
		t.Fatal(err)
	}
	if got := lines(result); got != "go.mod requires example.com/skipped, but nothing imports it" {
		t.Errorf("got:\n%s", got)
	}
}

func TestCompareTreatsWorkspaceMembersAsProvided(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.work", "go 1.22\n\nuse (\n\t./apps/a\n\t./libs/b\n)\n")
	write(t, root, "apps/a/go.mod", "module example.com/mono/apps/a\n\ngo 1.22\n")
	write(t, root, "apps/a/a.go", "package a\n\nimport _ \"example.com/mono/libs/b/deep\"\n")
	write(t, root, "libs/b/go.mod", "module example.com/mono/libs/b\n\ngo 1.22\n")
	write(t, root, "libs/b/b.go", "package b\n")

	result, leans, err := Compare(root, "apps/a")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Clean() || !leans {
		t.Errorf("a workspace member is provided by go.work: clean=%v leans=%v\n%s", result.Clean(), leans, lines(result))
	}
}

func TestToolPackagesReadsBlocksAndLines(t *testing.T) {
	got := toolPackages("module m\n\ntool example.com/one/cmd\n\ntool (\n\texample.com/two/cmd // note\n\texample.com/three\n)\n")
	want := []string{"example.com/one/cmd", "example.com/two/cmd", "example.com/three"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSupportsTidyDiff(t *testing.T) {
	for output, want := range map[string]bool{
		"go version go1.22.5 windows/amd64":     false,
		"go version go1.23.0 linux/amd64":       true,
		"go version go1.27.0 windows/amd64":     true,
		"go version go2.0 linux/amd64":          true,
		"go version devel go1.24-abc linux/arm": false,
		"go version go1.23rc1 darwin/arm64":     true,
		"nonsense":                              false,
	} {
		if got := supportsTidyDiff(output); got != want {
			t.Errorf("%q: got %v, want %v", output, got, want)
		}
	}
}
