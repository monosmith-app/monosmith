package apisurface

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListReadsTheWorkingTreeAndSkipsNestedModules(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, "libs", "core", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/core\n\ngo 1.22\n")
	write("core.go", "package core\nfunc Run() {}\n")
	write("nested/go.mod", "module example.com/core/nested\n")
	write("nested/n.go", "package nested\nfunc Other() {}\n")
	got, err := List(root, protocol.APISurfaceParams{Project: protocol.Project{Root: "libs/core"}})
	if err != nil || got.Unreadable != "" || !reflect.DeepEqual(got.Names, []string{"Run"}) {
		t.Fatalf("result %+v, error %v", got, err)
	}
	none, _ := List(root, protocol.APISurfaceParams{Project: protocol.Project{Root: "libs/missing"}})
	if none.Unreadable == "" || len(none.Names) != 0 {
		t.Fatalf("a project without go.mod is unreadable: %+v", none)
	}
}
