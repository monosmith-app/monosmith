package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declaredWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for name, body := range files {
		full := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestADeclaredBinFolderIsAProgramOfTheCrateAbove(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"crates/tools/Cargo.toml":          "[package]\nname = \"acme-tools\"\nversion = \"0.1.0\"\n\n[dependencies]\ncore = { path = \"../core\", package = \"acme-core\" }\n",
		"crates/tools/src/lib.rs":          "",
		"crates/tools/src/bin/gen/main.rs": "fn main() {}\n",
	})
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "gen", Root: "crates/tools/src/bin/gen"}})
	if err != nil || len(res.Projects) != 1 {
		t.Fatalf("%v %v", res.Projects, err)
	}
	p := res.Projects[0]
	if p.Language != "rust" || p.Kind != protocol.ProjectApp || !slices.Equal(p.Imports, []string{"acme-core"}) || len(p.Exports) != 0 {
		t.Errorf("%+v", p)
	}
	if got := strings.Join(p.Tasks["build"].Command, " "); got != "cargo build -p acme-tools --bin gen" {
		t.Errorf("build: %s", got)
	}
	if got := strings.Join(p.Tasks["test"].Command, " "); got != "cargo test -p acme-tools --bin gen" || p.Tasks["start"].Command == nil {
		t.Errorf("test: %s, start: %v", got, p.Tasks["start"])
	}
	if !slices.Contains(p.Tasks["test"].Inputs, "{workspaceRoot}/crates/tools/Cargo.toml") {
		t.Errorf("the crate's manifest is an input: %v", p.Tasks["test"].Inputs)
	}
}

func TestADeclaredRustFolderIsRefusedWhenItCannotBeBuilt(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"crates/tools/Cargo.toml":          "[package]\nname = \"acme-tools\"\n",
		"crates/tools/src/bin/gen/main.rs": "fn main() {}\n",
		"crates/tools/src/helpers/mod.rs":  "",
		"crates/own/Cargo.toml":            "[package]\nname = \"own\"\n",
		"loose/main.rs":                    "",
	})
	cases := map[string]struct {
		root, want string
		kind       protocol.ProjectKind
	}{
		"missing":     {"crates/tools/src/bin/none", "does not exist", ""},
		"own":         {"crates/own", "Cargo.toml of its own", ""},
		"no manifest": {"loose", "no Cargo.toml", ""},
		"not a bin":   {"crates/tools/src/helpers", "src/bin/<name>/main.rs", ""},
		"package":     {"crates/tools/src/bin/gen", "cannot be released on its own", protocol.ProjectPackage},
	}
	for name, c := range cases {
		_, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "x", Root: c.root, Kind: c.kind}})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
