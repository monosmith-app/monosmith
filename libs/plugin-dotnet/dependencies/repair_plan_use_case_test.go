package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairRestoresAgainRebuildsLocksAndClearsNuGet(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"apps/api", "libs/core"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "api", "packages.lock.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}})
	if err != nil || len(res.Plans) != 3 {
		t.Fatalf("reinstall, lockfile, cache: %+v, %v", res.Plans, err)
	}
	re, lock, cache := res.Plans[0], res.Plans[1], res.Plans[2]
	if re.Kind != protocol.RepairReinstall || len(re.Commands) != 2 || !reflect.DeepEqual(re.Commands[1], protocol.RepairCommand{Argv: []string{"dotnet", "restore", "--force"}, Dir: "libs/core"}) {
		t.Fatalf("reinstall: %+v", re)
	}
	if lock.Kind != protocol.RepairLockfile || !reflect.DeepEqual(lock.Delete, []string{"apps/api/packages.lock.json"}) || len(lock.Commands) != 1 || lock.Commands[0].Argv[2] != "--force-evaluate" {
		t.Fatalf("lockfile: %+v", lock)
	}
	if cache.Kind != protocol.RepairCache || !cache.MachineWide || !reflect.DeepEqual(cache.Commands[0].Argv, []string{"dotnet", "nuget", "locals", "all", "--clear"}) {
		t.Fatalf("cache: %+v", cache)
	}
}
