package dependencies_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// repairLab is a workspace with a root node_modules, one project that has its
// own node_modules and one that has none.
func repairLab(t *testing.T, withLock bool) (string, []protocol.Project) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"node_modules/left-pad", "apps/web/node_modules/.bin", "libs/ui/src"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if withLock {
		if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, []protocol.Project{
		{Name: "web", Root: "apps/web", Language: "node"},
		{Name: "ui", Root: "libs/ui", Language: "node"},
	}
}

func plansByKind(t *testing.T, res protocol.RepairResult) map[protocol.RepairKind]protocol.RepairPlan {
	t.Helper()
	out := map[protocol.RepairKind]protocol.RepairPlan{}
	for _, p := range res.Plans {
		if _, twice := out[p.Kind]; twice {
			t.Fatalf("kind %s has two plans", p.Kind)
		}
		out[p.Kind] = p
	}
	return out
}

func TestRepairOffersTheFourKindsLightestFirst(t *testing.T) {
	root, projects := repairLab(t, true)
	res, err := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []protocol.RepairKind
	for _, p := range res.Plans {
		kinds = append(kinds, p.Kind)
	}
	if !reflect.DeepEqual(kinds, protocol.RepairKinds()) {
		t.Fatalf("kinds = %v, want %v", kinds, protocol.RepairKinds())
	}
}

func TestRepairPruneRunsNpmPruneAndDeletesNothing(t *testing.T) {
	root, projects := repairLab(t, true)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	prune := plansByKind(t, res)[protocol.RepairPrune]
	if len(prune.Delete) != 0 || prune.Install || prune.MachineWide || prune.ChangesVersions {
		t.Fatalf("prune only runs npm prune: %+v", prune)
	}
	want := []string{"npm", "prune", "--no-audit", "--no-fund"}
	if len(prune.Commands) != 1 || !reflect.DeepEqual(prune.Commands[0].Argv, want) {
		t.Fatalf("prune commands = %+v, want %v", prune.Commands, want)
	}
}

func TestRepairReinstallDeletesEveryNodeModulesThatExistsAndKeepsTheLockfile(t *testing.T) {
	root, projects := repairLab(t, true)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	reinstall := plansByKind(t, res)[protocol.RepairReinstall]
	want := []string{"node_modules", "apps/web/node_modules"}
	if !reflect.DeepEqual(reinstall.Delete, want) {
		t.Fatalf("reinstall deletes %v, want %v (libs/ui has no node_modules)", reinstall.Delete, want)
	}
	if !reinstall.Install || reinstall.ChangesVersions || reinstall.MachineWide || len(reinstall.Commands) != 0 {
		t.Fatalf("reinstall deletes and installs, nothing else: %+v", reinstall)
	}
}

func TestRepairLockfileAlsoDeletesTheLockfileAndSaysVersionsCanChange(t *testing.T) {
	root, projects := repairLab(t, true)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	lock := plansByKind(t, res)[protocol.RepairLockfile]
	want := []string{"node_modules", "apps/web/node_modules", "package-lock.json"}
	if !reflect.DeepEqual(lock.Delete, want) {
		t.Fatalf("lockfile deletes %v, want %v", lock.Delete, want)
	}
	if !lock.Install || !lock.ChangesVersions || lock.MachineWide {
		t.Fatalf("lockfile installs and can change versions: %+v", lock)
	}
}

func TestRepairLeavesTheLockfileKindOutWhenThereIsNoLockfile(t *testing.T) {
	root, projects := repairLab(t, false)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	if _, ok := plansByKind(t, res)[protocol.RepairLockfile]; ok {
		t.Fatal("a workspace without package-lock.json has no lockfile to rebuild")
	}
	if len(res.Plans) != 3 {
		t.Fatalf("want prune, reinstall and cache, got %d plans", len(res.Plans))
	}
}

func TestRepairCacheSaysItIsMachineWide(t *testing.T) {
	root, projects := repairLab(t, true)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects})
	cache := plansByKind(t, res)[protocol.RepairCache]
	want := []string{"npm", "cache", "clean", "--force"}
	if len(cache.Commands) != 1 || !reflect.DeepEqual(cache.Commands[0].Argv, want) {
		t.Fatalf("cache commands = %+v, want %v", cache.Commands, want)
	}
	if !cache.MachineWide || !cache.Install || len(cache.Delete) != 0 || len(cache.Notes) == 0 {
		t.Fatalf("cache is machine-wide, installs after, deletes nothing in the workspace and says whose cache it is: %+v", cache)
	}
}

func TestRepairForOneProjectStillNamesTheRootNodeModules(t *testing.T) {
	root, projects := repairLab(t, true)
	res, _ := dependencies.Repair(root, protocol.RepairParams{Projects: projects[1:]})
	reinstall := plansByKind(t, res)[protocol.RepairReinstall]
	if !reflect.DeepEqual(reinstall.Delete, []string{"node_modules"}) {
		t.Fatalf("npm installs the workspace as one tree: want the root node_modules only, got %v", reinstall.Delete)
	}
}

func TestRepairHasNoPlanWithoutProjects(t *testing.T) {
	res, err := dependencies.Repair(t.TempDir(), protocol.RepairParams{})
	if err != nil || len(res.Plans) != 0 {
		t.Fatalf("no project, no plan: %+v, %v", res, err)
	}
}

func TestRepairChangesNothingOnDisk(t *testing.T) {
	root, projects := repairLab(t, true)
	if _, err := dependencies.Repair(root, protocol.RepairParams{Projects: projects}); err != nil {
		t.Fatal(err)
	}
	for _, kept := range []string{"node_modules/left-pad", "apps/web/node_modules/.bin", "package-lock.json"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(kept))); err != nil {
			t.Fatalf("Repair only describes, but %s is gone: %v", kept, err)
		}
	}
}
