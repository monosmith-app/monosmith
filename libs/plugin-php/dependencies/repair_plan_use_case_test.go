package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairDeletesVendorAndComposerLock(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "vendor", "acme"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composer.lock"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "site", Root: "apps/site"}}})
	if err != nil || len(res.Plans) != 3 {
		t.Fatalf("reinstall, lockfile, cache: %+v, %v", res.Plans, err)
	}
	if !reflect.DeepEqual(res.Plans[0].Delete, []string{"vendor"}) || !res.Plans[0].Install {
		t.Fatalf("reinstall: %+v", res.Plans[0])
	}
	if !reflect.DeepEqual(res.Plans[1].Delete, []string{"vendor", "composer.lock"}) || !res.Plans[1].ChangesVersions {
		t.Fatalf("lockfile: %+v", res.Plans[1])
	}
	if res.Plans[2].Commands[0].Argv[1] != "clear-cache" || !res.Plans[2].MachineWide {
		t.Fatalf("cache: %+v", res.Plans[2])
	}
}
