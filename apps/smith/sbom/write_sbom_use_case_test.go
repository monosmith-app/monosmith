package sbom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoSBOMListsEveryRequiredModule(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "app"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "app", "go.mod"), []byte("module example.com/app\n\ngo 1.22\n\nrequire github.com/a/b v1.2.3\n\nrequire (\n\tgolang.org/x/sys v0.1.0 // indirect\n)\n"), 0o644)
	rel, err := Write(Request{Root: root, ProjectRoot: "app", Language: "go", Name: "app", Version: "1.0.0", OutDir: "dist/release/app", SmithVersion: "x"})
	if err != nil || rel != "dist/release/app/app_1.0.0_sbom.cdx.json" {
		t.Fatalf("%q %v", rel, err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	var doc struct {
		BOMFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
		Components  []struct{ PURL string }
	}
	if err := json.Unmarshal(raw, &doc); err != nil || doc.BOMFormat != "CycloneDX" || doc.SpecVersion != "1.5" || len(doc.Components) != 2 {
		t.Fatalf("%s %v", raw, err)
	}
	if doc.Components[1].PURL != "pkg:golang/golang.org/x/sys@v0.1.0" || !strings.Contains(string(raw), "pkg:golang/example.com/app@1.0.0") {
		t.Errorf("purls: %s", raw)
	}
}

func TestNPMSBOMReadsTheWorkspaceLockfile(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "package-lock.json"), []byte(`{"packages":{"":{},"pkg":{"version":"1.0.0"},"node_modules/pkg":{"link":true},"node_modules/left-pad":{"version":"1.3.0"},"node_modules/@scope/x":{"version":"2.0.0"},"node_modules/a/node_modules/left-pad":{"version":"1.3.0"}}}`), 0o644)
	comps, err := nodeComponents(Request{Root: root, ProjectRoot: "pkg"})
	if err != nil || len(comps) != 2 {
		t.Fatalf("%v %v", comps, err)
	}
	if comps[0].PURL != "pkg:npm/%40scope/x@2.0.0" || comps[1].PURL != "pkg:npm/left-pad@1.3.0" {
		t.Errorf("%v", comps)
	}
}

func TestNPMFallsBackToTheManifestAndOtherLanguagesGetNothing(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"x":"^1.0.0"}}`), 0o644)
	comps, err := nodeComponents(Request{Root: root, ProjectRoot: "."})
	if err != nil || len(comps) != 1 || comps[0].PURL != "pkg:npm/x@^1.0.0" {
		t.Fatalf("%v %v", comps, err)
	}
	if rel, err := Write(Request{Root: root, Language: "rust"}); rel != "" || err != nil {
		t.Errorf("%q %v", rel, err)
	}
}
