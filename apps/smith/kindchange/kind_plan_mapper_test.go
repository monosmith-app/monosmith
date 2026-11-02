package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestOperationPlanShowsTheMoveTheEditsAndTheirDiffs(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{\n  \"private\": true\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan := OperationPlan(root, Plan{
		Project: "core", From: protocol.ProjectLib, To: protocol.ProjectPackage,
		Root: "libs/core", NewRoot: "packages/core",
		Edits: []protocol.FileEdit{
			{Path: "package.json", Content: "{\n  \"private\": false\n}\n"},
			{Path: "old.txt", Delete: true},
			{Path: "fresh.txt", Content: "x\n"},
		},
	})
	var b strings.Builder
	operationplan.Render(&b, plan)
	out := b.String()
	for _, want := range []string{"make core a package (it is a lib)", "libs/core -> packages/core", "-  \"private\": true", "+  \"private\": false", "remove  old.txt", "create  fresh.txt", "CI pipelines"} {
		if !strings.Contains(out, want) {
			t.Errorf("the plan lacks %q:\n%s", want, out)
		}
	}
}

func TestDirtyPaths(t *testing.T) {
	for _, tc := range []struct {
		name, status string
		want         int
	}{
		{"clean", "", 0},
		{"smith's own folder does not count", "?? .smith/\n M .smith/cache.json\n", 0},
		{"modified and untracked count", " M src/a.go\n?? notes.txt\n", 2},
		{"a rename counts once", "R  a.go -> b.go\n", 1},
	} {
		if got := DirtyPaths(tc.status); len(got) != tc.want {
			t.Errorf("%s: %v", tc.name, got)
		}
	}
}
