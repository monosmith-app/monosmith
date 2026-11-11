package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func goModule(t *testing.T, root, dir string, sum bool) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(dir))
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(full, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if sum {
		if err := os.WriteFile(filepath.Join(full, "go.sum"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRepairTidiesRebuildsGoSumAndClearsTheModuleCache(t *testing.T) {
	root := t.TempDir()
	goModule(t, root, "apps/api", true)
	goModule(t, root, "libs/core", false)
	res, err := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []protocol.RepairKind
	for _, p := range res.Plans {
		kinds = append(kinds, p.Kind)
	}
	if !reflect.DeepEqual(kinds, []protocol.RepairKind{protocol.RepairPrune, protocol.RepairLockfile, protocol.RepairCache}) {
		t.Fatalf("kinds %v", kinds)
	}
	prune, lock, cache := res.Plans[0], res.Plans[1], res.Plans[2]
	if len(prune.Commands) != 2 || prune.Commands[0].Dir != "apps/api" || !reflect.DeepEqual(prune.Commands[0].Argv, []string{"go", "mod", "tidy"}) {
		t.Fatalf("prune tidies each module: %+v", prune.Commands)
	}
	if !reflect.DeepEqual(lock.Delete, []string{"apps/api/go.sum"}) {
		t.Fatalf("lockfile deletes the go.sum that exists: %v", lock.Delete)
	}
	if !cache.MachineWide || !cache.Install || !reflect.DeepEqual(cache.Commands[0].Argv, []string{"go", "clean", "-modcache"}) {
		t.Fatalf("cache: %+v", cache)
	}
}

func TestRepairWithoutGoSumHasNoLockfileKind(t *testing.T) {
	root := t.TempDir()
	goModule(t, root, ".", false)
	res, _ := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "app", Root: "."}}})
	for _, p := range res.Plans {
		if p.Kind == protocol.RepairLockfile {
			t.Fatal("no go.sum, no lockfile kind")
		}
	}
	if res.Plans[0].Commands[0].Dir != "" {
		t.Fatalf("a module at the root runs at the root: %+v", res.Plans[0].Commands[0])
	}
}
