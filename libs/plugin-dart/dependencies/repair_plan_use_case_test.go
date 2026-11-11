package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairDeletesDartToolAndPubspecLock(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{".dart_tool", "apps/mobile/.dart_tool", "libs/core"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "pubspec.lock"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	projects := []protocol.Project{
		{Name: "mobile", Root: "apps/mobile", Tasks: map[string]protocol.Task{"build": {Command: []string{"flutter", "build"}}}},
		{Name: "core", Root: "libs/core"},
	}
	res, err := Repair(root, protocol.RepairParams{Projects: projects})
	if err != nil || len(res.Plans) != 3 {
		t.Fatalf("reinstall, lockfile, cache: %+v, %v", res.Plans, err)
	}
	if want := []string{".dart_tool", "apps/mobile/.dart_tool"}; !reflect.DeepEqual(res.Plans[0].Delete, want) || !res.Plans[0].Install {
		t.Fatalf("reinstall: %+v", res.Plans[0])
	}
	if want := []string{".dart_tool", "apps/mobile/.dart_tool", "pubspec.lock"}; !reflect.DeepEqual(res.Plans[1].Delete, want) || !res.Plans[1].ChangesVersions {
		t.Fatalf("lockfile: %+v", res.Plans[1])
	}
	if want := []string{"flutter", "pub", "cache", "clean", "--force"}; !reflect.DeepEqual(res.Plans[2].Commands[0].Argv, want) || !res.Plans[2].MachineWide {
		t.Fatalf("cache uses flutter in a Flutter workspace: %+v", res.Plans[2])
	}
}
