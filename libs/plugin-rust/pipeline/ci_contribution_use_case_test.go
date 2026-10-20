package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIAddsTheRustPolicyStepsOnlyWhenOptedIn(t *testing.T) {
	projects := []protocol.Project{{Name: "cli", Root: "crates/cli"}}
	off := CI(t.TempDir(), protocol.CIParams{Projects: projects})
	if len(off.Audit) != 0 || len(off.Setup) != 1 {
		t.Fatalf("off by default: %+v", off)
	}
	on := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"rust"}, MaxUnsafe: 3}})
	if len(on.Setup) != 2 || on.Setup[1].Command[3] != "cargo-deny" || on.Setup[1].Command[5] != CargoDenyVersion {
		t.Fatalf("setup: %+v", on.Setup)
	}
	wantDeny := protocol.Step{Name: "cargo deny (licenses, bans)", Command: []string{"cargo", "deny", "check", "licenses", "bans"}, Blocking: true, CacheFiles: []string{"Cargo.lock", "Cargo.toml", "deny.toml", "crates/cli/Cargo.toml"}}
	wantUnsafe := protocol.Step{Name: "rust unsafe count", Command: []string{"{smith}", "rust-policy", "--max-unsafe=3", "crates/cli"}, Blocking: true}
	if len(on.Audit) != 2 || !reflect.DeepEqual(on.Audit[0], wantDeny) || !reflect.DeepEqual(on.Audit[1], wantUnsafe) {
		t.Fatalf("audit: %+v", on.Audit)
	}
	warn := CI(t.TempDir(), protocol.CIParams{Projects: projects, Audit: protocol.AuditSettings{Policies: []string{"rust"}, Warn: []string{"rust"}}})
	if warn.Audit[0].Blocking || warn.Audit[1].Blocking {
		t.Error("warn makes both report only")
	}
}

func TestCIDeclaresTheCargoRegistryStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "cli", Root: "crates/cli"}}}).Stores
	want := []protocol.PackageStore{{Name: "cargo", Path: "~/.cargo/registry", Env: "CARGO_HOME", Files: []string{"**/Cargo.lock"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
