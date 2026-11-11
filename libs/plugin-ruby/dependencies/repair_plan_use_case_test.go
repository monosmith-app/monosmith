package dependencies

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairRedownloadsAndRebuildsGemfileLock(t *testing.T) {
	root := t.TempDir()
	projects := protocol.RepairParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}}
	res, _ := Repair(root, projects)
	if len(res.Plans) != 1 || !reflect.DeepEqual(res.Plans[0].Commands[0].Argv, []string{"bundle", "install", "--redownload"}) {
		t.Fatalf("without Gemfile.lock only reinstall: %+v", res.Plans)
	}
	if err := os.WriteFile(filepath.Join(root, "Gemfile.lock"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	res, _ = Repair(root, projects)
	if len(res.Plans) != 2 || res.Plans[1].Kind != protocol.RepairLockfile || res.Plans[1].Delete[0] != "Gemfile.lock" || !res.Plans[1].ChangesVersions {
		t.Fatalf("lockfile: %+v", res.Plans)
	}
}
