package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListReadsTheWorkspaceLockfile(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "package-lock.json"), []byte(`{"packages":{"":{},"pkg":{"version":"1.0.0"},"node_modules/pkg":{"link":true},"node_modules/left-pad":{"version":"1.3.0"},"node_modules/@scope/x":{"version":"2.0.0"},"node_modules/a/node_modules/left-pad":{"version":"1.3.0"}}}`), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "pkg"}})
	if err != nil || len(res.Components) != 2 {
		t.Fatalf("%+v %v", res, err)
	}
	if res.Components[0].PURL != "pkg:npm/%40scope/x@2.0.0" || res.Components[1].PURL != "pkg:npm/left-pad@1.3.0" {
		t.Errorf("%v", res.Components)
	}
}

func TestListFallsBackToThePackageJSON(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"x":"^1.0.0"}}`), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].PURL != "pkg:npm/x@^1.0.0" {
		t.Fatalf("%+v %v", res, err)
	}
}
