package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairResetsResolvesAndPurgesTheCache(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "apps", "tool"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "tool", "Package.resolved"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "tool", Root: "apps/tool"}}})
	if err != nil || len(res.Plans) != 3 {
		t.Fatalf("reinstall, lockfile, cache: %+v, %v", res.Plans, err)
	}
	re := res.Plans[0]
	if len(re.Commands) != 2 || !reflect.DeepEqual(re.Commands[0].Argv, []string{"swift", "package", "reset"}) || re.Commands[0].Dir != "apps/tool" || re.Install {
		t.Fatalf("reinstall resets and resolves, the plugin cannot install: %+v", re)
	}
	if !reflect.DeepEqual(res.Plans[1].Delete, []string{"apps/tool/Package.resolved"}) || !res.Plans[1].ChangesVersions {
		t.Fatalf("lockfile: %+v", res.Plans[1])
	}
	if c := res.Plans[2]; !c.MachineWide || c.Install || c.Commands[0].Argv[2] != "purge-cache" || len(c.Commands) != 2 {
		t.Fatalf("cache: %+v", c)
	}
}
