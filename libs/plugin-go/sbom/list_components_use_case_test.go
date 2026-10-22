package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListEveryRequiredModule(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "app"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "app", "go.mod"), []byte("module example.com/app\n\ngo 1.22\n\nrequire github.com/a/b v1.2.3\n\nrequire (\n\tgolang.org/x/sys v0.1.0 // indirect\n)\n"), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || len(res.Components) != 2 || res.PURL != "pkg:golang/example.com/app" || res.Source != "app/go.mod" {
		t.Fatalf("%+v %v", res, err)
	}
	if res.Components[1].PURL != "pkg:golang/golang.org/x/sys@v0.1.0" {
		t.Errorf("%v", res.Components)
	}
}

func TestListWithoutGoModFails(t *testing.T) {
	if _, err := List(t.TempDir(), protocol.ComponentsParams{Project: protocol.Project{Root: "."}}); err == nil {
		t.Error("a Go project needs a go.mod")
	}
}
