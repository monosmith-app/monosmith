package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

func TestPlanRemovalListsTheDirectoryThenTheWorkspaceFilesWithOwners(t *testing.T) {
	root := t.TempDir()
	lock := "schema = 1\n\n[[file]]\npath = \"smith.toml\"\nowner = \"you\"\nrenderer = \"r\"\nversion = 1\nmode = \"file\"\nhash = \"h\"\nwanted = \"h\"\n"
	if err := os.WriteFile(filepath.Join(root, "smith.lock"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	plan := PlanRemoval(root, Removal{Project: "web", Root: "apps/web", Files: []string{"smith.toml", ".github/workflows/ci.yml"}})
	if plan.Title != "remove web" || len(plan.Groups) != 2 {
		t.Fatalf("%+v", plan)
	}
	if l := plan.Groups[0].Lines[0]; l.Action != operationplan.Remove || l.Path != "apps/web/" {
		t.Fatalf("%+v", l)
	}
	files := plan.Groups[1].Lines
	if files[0].Owner != "you" || files[1].Owner != "" {
		t.Fatalf("%+v", files)
	}
}
