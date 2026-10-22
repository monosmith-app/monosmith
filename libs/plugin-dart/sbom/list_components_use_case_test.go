package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestListReadsPubspecLockAndSkipsSDKAndDev(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "pubspec.lock"), []byte(`packages:
  http:
    dependency: "direct main"
    source: hosted
    version: "1.2.0"
  flutter:
    dependency: "direct main"
    source: sdk
    version: "0.0.0"
  lints:
    dependency: "direct dev"
    source: hosted
    version: "4.0.0"
`), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].PURL != "pkg:pub/http@1.2.0" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToPubspec(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "pubspec.yaml"), []byte("name: x\ndependencies:\n  flutter:\n    sdk: flutter\n  path: ^1.9.0\n  git_dep:\n    git: https://example.com/x.git\n"), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 1 || res.Components[0].Name != "path" {
		t.Fatalf("%+v %v", res, err)
	}
}
