package pipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIAddsTheRustPolicyStepsOnlyWhenOptedIn(t *testing.T) {
	projects := []protocol.Project{{Name: "cli", Root: "crates/cli"}}
	off := CI(t.TempDir(), protocol.CIParams{Projects: projects})
	if len(off.Audit) != 1 || off.Audit[0].Name != "cargo audit" || off.Audit[0].Blocking || len(off.Setup) != 1 {
		t.Fatalf("off by default: %+v", off)
	}
	on := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"rust"}, MaxUnsafe: 3}})
	if len(on.Setup) != 2 || on.Setup[1].Command[3] != "cargo-deny" || on.Setup[1].Command[5] != CargoDenyVersion {
		t.Fatalf("setup: %+v", on.Setup)
	}
	wantDeny := protocol.Step{Name: "cargo deny (licenses, bans)", Command: []string{"cargo", "deny", "check", "licenses", "bans"}, Blocking: true, CacheFiles: []string{"Cargo.lock", "Cargo.toml", "deny.toml", "crates/cli/Cargo.toml"}}
	wantUnsafe := protocol.Step{Name: "rust unsafe count", Command: []string{"{smith}", "rust-policy", "--max-unsafe=3", "crates/cli"}, Blocking: true}
	if len(on.Audit) != 3 || !reflect.DeepEqual(on.Audit[1], wantDeny) || !reflect.DeepEqual(on.Audit[2], wantUnsafe) {
		t.Fatalf("audit: %+v", on.Audit)
	}
	warn := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"rust"}, Warn: []string{"rust"}}})
	if warn.Audit[1].Blocking || warn.Audit[2].Blocking {
		t.Error("warn makes both report only")
	}
}

func TestCIAlwaysAuditsCratesForKnownVulnerabilities(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "cli", Root: "crates/cli"}}}).Audit
	if len(got) != 1 || got[0].Blocking {
		t.Fatalf("one report-only step: %+v", got)
	}
	if want := []string{"Cargo.lock", "Cargo.toml", "crates/cli/Cargo.toml"}; !reflect.DeepEqual(got[0].CacheFiles, want) {
		t.Errorf("the verdict is remembered per lockfile and manifest: %v", got[0].CacheFiles)
	}
	script := got[0].Command[2]
	for _, want := range []string{"cargo audit", "cargo install --locked cargo-audit", `"$CI"`} {
		if !strings.Contains(script, want) {
			t.Errorf("the step says how to get the tool (%q missing): %s", want, script)
		}
	}
}

func TestCIDeclaresTheCargoRegistryStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "cli", Root: "crates/cli"}}}).Stores
	want := []protocol.PackageStore{{Name: "cargo", Path: "~/.cargo/registry", Env: "CARGO_HOME", Files: []string{"**/Cargo.lock"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
