package dependencies

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairRebuildsCargoLockOnly(t *testing.T) {
	root := t.TempDir()
	projects := protocol.RepairParams{Projects: []protocol.Project{{Name: "cli", Root: "apps/cli"}}}
	if res, _ := Repair(root, projects); len(res.Plans) != 0 {
		t.Fatalf("no Cargo.lock, nothing to repair: %+v", res.Plans)
	}
	if err := os.WriteFile(filepath.Join(root, "Cargo.lock"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Repair(root, projects)
	if err != nil || len(res.Plans) != 1 {
		t.Fatalf("one plan: %+v, %v", res.Plans, err)
	}
	p := res.Plans[0]
	if p.Kind != protocol.RepairLockfile || p.Delete[0] != "Cargo.lock" || !p.Install || !p.ChangesVersions {
		t.Fatalf("lockfile: %+v", p)
	}
}
