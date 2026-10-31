package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListReadsVcpkgAndConan(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "vcpkg.json"), []byte(`{"dependencies":["fmt",{"name":"zlib","version>=":"1.3"},{"name":"vcpkg-cmake","host":true}],"overrides":[{"name":"fmt","version":"10.2.1"}]}`), 0o644)
	_ = os.WriteFile(filepath.Join(root, "conanfile.txt"), []byte("[requires]\nboost/1.84.0@user/stable\n\n[generators]\nCMakeDeps\n"), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 3 || res.Source != "vcpkg.json" {
		t.Fatalf("%+v %v", res, err)
	}
	want := map[string]bool{"pkg:generic/fmt@10.2.1": true, "pkg:generic/zlib@1.3": true, "pkg:conan/boost@1.84.0": true}
	for _, c := range res.Components {
		if !want[c.PURL] {
			t.Errorf("unexpected %s", c.PURL)
		}
	}
}

func TestListWithNoDeclaredLibrariesIsEmpty(t *testing.T) {
	res, err := List(t.TempDir(), protocol.ComponentsParams{Project: protocol.Project{Root: "app"}})
	if err != nil || len(res.Components) != 0 || res.Source != "" {
		t.Fatalf("%+v %v", res, err)
	}
}
