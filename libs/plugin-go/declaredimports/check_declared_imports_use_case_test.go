package declaredimports

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lines(found []UndeclaredImport) []string {
	out := []string{}
	for _, u := range found {
		out = append(out, u.String())
	}
	return out
}

const goMod = `module example.com/acme/ocr

go 1.26

require github.com/spf13/cobra v1.10.2 // indirect

require (
	golang.org/x/text v0.30.0
	example.com/acme/legacy v0.0.0
)

replace example.com/acme/legacy => ../legacy
`

func TestCheckReportsImportsNoRequireProvides(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "go.work", "go 1.26\n\nuse (\n\t./libs/core\n\t./packages/ocr\n)\n")
	write(t, ws, "libs/core/go.mod", "module example.com/acme/core\n\ngo 1.26\n")
	write(t, ws, "packages/ocr/go.mod", goMod)
	write(t, ws, "packages/ocr/read.go", `package ocr

import (
	"fmt"
	"net/http"

	"example.com/acme/core"
	"example.com/acme/legacy/old"
	"example.com/acme/ocr/internal/pages"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
	"C"
)
`)
	write(t, ws, "packages/ocr/read_test.go", "package ocr_test\n\nimport _ \"github.com/google/go-cmp/cmp\"\n")
	write(t, ws, "packages/ocr/gen.go", "//go:build ignore\n\npackage main\n\nimport _ \"golang.org/x/tools/go/packages\"\n")
	write(t, ws, "packages/ocr/vendor/x/x.go", "package x\n\nimport _ \"github.com/vendored/thing\"\n")
	write(t, ws, "packages/ocr/testdata/x.go", "package x\n\nimport _ \"github.com/fixture/thing\"\n")
	write(t, ws, "packages/ocr/cmd/tool/go.mod", "module example.com/acme/ocr/cmd/tool\n")
	write(t, ws, "packages/ocr/cmd/tool/main.go", "package main\n\nimport _ \"github.com/nested/module\"\n")
	write(t, ws, "packages/ocr/broken.go", "package ocr\n\nimport (\n")

	got, err := Check(ws, "packages/ocr")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"read.go imports github.com/stretchr/testify/require, which go.mod does not require",
		"read.go imports golang.org/x/sys/unix, which go.mod does not require",
		"read.go imports gopkg.in/yaml.v3, which go.mod does not require",
		"read_test.go imports github.com/google/go-cmp/cmp, which go.mod does not require",
	}
	if !reflect.DeepEqual(lines(got), want) {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(lines(got), "\n"), strings.Join(want, "\n"))
	}
}

func TestCheckNeedsAModuleLine(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "m/go.mod", "go 1.26\n")
	if _, err := Check(ws, "m"); err == nil || !strings.Contains(err.Error(), "no module line") {
		t.Fatalf("err = %v", err)
	}
	if _, err := Check(ws, "missing"); err == nil {
		t.Fatal("a directory without go.mod is an error")
	}
}
