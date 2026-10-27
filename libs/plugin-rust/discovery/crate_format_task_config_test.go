package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFormatRunsRustfmtOnSmithsDefaultUnlessTheRepositoryConfiguresIt(t *testing.T) {
	def := Tasks("core", protocol.ProjectLib, []string{"--config-path", "{workspaceRoot}/.smith/defaults/rustfmt.toml"})["format"]
	if got := strings.Join(def.Command, " "); got != "cargo fmt -p core --check -- --config-path {workspaceRoot}/.smith/defaults/rustfmt.toml" {
		t.Fatalf("command: %s", got)
	}
	if last := def.Inputs[len(def.Inputs)-1]; last != "{workspaceRoot}/.smith/defaults/rustfmt.toml" {
		t.Fatalf("the default's content is an input: %v", def.Inputs)
	}
	own := Tasks("core", protocol.ProjectLib, nil)["format"]
	if got := strings.Join(own.Command, " "); got != "cargo fmt -p core --check" {
		t.Fatalf("command: %s", got)
	}
	if strings.Contains(own.Description, "smith's default") {
		t.Fatalf("description: %s", own.Description)
	}
}

func TestDiscoverPassesTheDefaultRustfmtConfigToEachCrate(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "crates/core/Cargo.toml", "[package]\nname = \"core\"\nversion = \"0.1.0\"\n")
	write(t, ws, "crates/own/Cargo.toml", "[package]\nname = \"own\"\nversion = \"0.1.0\"\n")
	write(t, ws, "crates/own/rustfmt.toml", "tab_spaces = 2\n")
	res, err := Discover(ws, []string{"crates/core/Cargo.toml", "crates/own/Cargo.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(res.Projects[0].Tasks["format"].Command, " "); !strings.HasSuffix(got, "-- --config-path {workspaceRoot}/.smith/defaults/rustfmt.toml") {
		t.Errorf("core: %s", got)
	}
	if got := strings.Join(res.Projects[1].Tasks["format"].Command, " "); strings.Contains(got, "--config-path") {
		t.Errorf("own: %s", got)
	}
}
