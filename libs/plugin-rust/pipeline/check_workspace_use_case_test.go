package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctor(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "Cargo.toml"), []byte("[workspace]\nmembers = [\"apps/cli\", \"libs/gone\", \"crates/*\"]\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "cli", "Cargo.toml"), []byte("[package]\nname = \"cli\"\n"), 0o644)
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("not found") }
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/cli"}, {Root: "crates/x"}, {Root: "libs/new"}}})
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.Check] = f.OK
	}
	want := map[string]bool{
		"cargo is installed": false, "workspace member apps/cli exists": true, "workspace member libs/gone exists": false,
		"the Cargo workspace lists apps/cli": true, "the Cargo workspace lists crates/x": true, "the Cargo workspace lists libs/new": false,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %v, want %v", k, got[k], v)
		}
	}
}

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}})
	if res.Toolchains[0].Version != RustVersion || len(res.Credentials) != 1 || res.Setup[0].Command[1] != "fetch" {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, "rust-toolchain.toml"), []byte("[toolchain]\nchannel = \"1.85.0\"\n"), 0o644)
	if v := CI(ws, protocol.CIParams{Projects: []protocol.Project{{}}}).Toolchains[0].Version; v != "1.85.0" {
		t.Fatalf("rust-toolchain.toml pins %s", v)
	}
}

func TestDoctorCrateWithItsOwnWorkspaceOrExcludeIsNotMissing(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, content string) {
		_ = os.MkdirAll(filepath.Dir(filepath.Join(ws, rel)), 0o755)
		_ = os.WriteFile(filepath.Join(ws, rel), []byte(content), 0o644)
	}
	write("Cargo.toml", "[workspace]\nmembers = [\"crates/a\"]\nexclude = [\"tools/old\"]\n")
	write("crates/a/Cargo.toml", "[package]\nname = \"a\"\n")
	write("tools/gpu-probe/Cargo.toml", "[package]\nname = \"gpu-probe\"\n\n[workspace]\n")
	write("tools/old/Cargo.toml", "[package]\nname = \"old\"\n")
	write("tools/new/Cargo.toml", "[package]\nname = \"new\"\n")
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("not found") }
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "crates/a"}, {Root: "tools/gpu-probe"}, {Root: "tools/old"}, {Root: "tools/new"}}})
	got := map[string]protocol.Finding{}
	for _, f := range res.Findings {
		got[f.Check] = f
	}
	for _, check := range []string{"tools/gpu-probe is its own Cargo workspace", "tools/old is excluded from the Cargo workspace", "the Cargo workspace lists crates/a"} {
		if f, ok := got[check]; !ok || !f.OK {
			t.Errorf("%q: %+v", check, f)
		}
	}
	if _, bad := got["the Cargo workspace lists tools/gpu-probe"]; bad {
		t.Error("a crate with its own [workspace] must not be reported as missing from members")
	}
	missing := got["the Cargo workspace lists tools/new"]
	if missing.OK || missing.Fix == nil || len(missing.Fix.Choices) != 1 || len(missing.Fix.Choices[0].Options) != 2 {
		t.Fatalf("an unknown crate must offer add-to-members or exclude: %+v", missing)
	}
	opts := missing.Fix.Choices[0].Options
	if c := opts[0].Fix.Files[0].Content; !strings.Contains(c, `"tools/new"`) || !strings.Contains(c, "members = [\"crates/a\", \"tools/new\"]") {
		t.Errorf("members option: %q", c)
	}
	if c := opts[1].Fix.Files[0].Content; !strings.Contains(c, "exclude = [\"tools/old\", \"tools/new\"]") {
		t.Errorf("exclude option: %q", c)
	}
	if audit := got["cargo-audit is installed"]; audit.OK {
		t.Log("cargo-audit finding present only when cargo is installed")
	}
}
