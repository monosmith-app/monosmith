package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIDeclaresTheCargoRegistryStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "cli", Root: "crates/cli"}}}).Stores
	want := []protocol.PackageStore{{Name: "cargo", Path: "~/.cargo/registry", Env: "CARGO_HOME", Files: []string{"**/Cargo.lock"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
}
