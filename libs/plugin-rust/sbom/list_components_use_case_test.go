package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const lock = `version = 3

[[package]]
name = "app"
version = "0.1.0"
dependencies = ["serde"]

[[package]]
name = "serde"
version = "1.0.200"
source = "registry+https://github.com/rust-lang/crates.io-index"
checksum = "abc"
`

func TestListReadsTheWorkspaceCargoLock(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "crates", "app"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "Cargo.lock"), []byte(lock), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "crates/app"}})
	if err != nil || res.Source != "Cargo.lock" || len(res.Components) != 1 || res.Components[0].PURL != "pkg:cargo/serde@1.0.200" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToTheManifest(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte("[package]\nname = \"app\"\n\n[dependencies]\nserde = \"1\"\nlocal = { path = \"../local\" }\ntokio = { version = \"1.3\", features = [\"rt\"] }\n"), 0o644)
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || res.Source != "Cargo.toml" || len(res.Components) != 2 || res.Components[0].PURL != "pkg:cargo/serde@1" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListWithNoFilesIsEmpty(t *testing.T) {
	res, err := List(t.TempDir(), protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || len(res.Components) != 0 || res.Source != "" {
		t.Fatalf("%+v %v", res, err)
	}
}
