package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("Cargo.toml", "[workspace]\nresolver = \"2\"\nmembers = [\"apps/cli\", \"libs/core\"]\n")
	write("apps/cli/Cargo.toml", "[package]\nname = \"acme-cli\"\nversion = \"0.3.0\"\nedition = \"2021\"\n\n[dependencies]\ncore = { path = \"../../libs/core\", package = \"acme-core\" }\nserde = \"1\"\n\n[dev-dependencies]\nhelpers = { path = \"../../libs/helpers\" }\n")
	write("apps/cli/src/main.rs", "fn main() {}\n")
	write("libs/core/Cargo.toml", "[package]\nname = \"acme-core\"\nversion = { workspace = true }\n\n[package.metadata.smith]\nkind = \"package\"\nstack = \"rust-strict\"\n")
	write("libs/core/src/lib.rs", "")
	res, err := Discover(ws, []string{"Cargo.toml", "apps/cli/Cargo.toml", "libs/core/Cargo.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("the virtual root is not a project: %+v", res.Projects)
	}
	cli, core := res.Projects[0], res.Projects[1]
	if cli.Name != "cli" || cli.Kind != protocol.ProjectApp || cli.Version != "0.3.0" || cli.Exports[0] != "acme-cli" || cli.PackageName != "acme-cli" || !reflect.DeepEqual(cli.Imports, []string{"acme-core", "helpers"}) {
		t.Errorf("cli %+v", cli)
	}
	if cli.Tasks["lint"].Command[1] != "clippy" || cli.Tasks["build"].Command[3] != "acme-cli" || cli.Tasks["start"].Command[1] != "run" {
		t.Errorf("cli tasks %+v", cli.Tasks)
	}
	if core.Kind != protocol.ProjectPackage || core.Stack != "rust-strict" || core.Version != "" || core.Tasks["package"].Command[1] != "package" {
		t.Errorf("core %+v", core)
	}
}
