package doctor

import (
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func lf(b []byte) string { return strings.ReplaceAll(string(b), "\r\n", "\n") }

func TestAnOldRepoGetsOneFindingWhoseFixRenamesTheKeys(t *testing.T) {
	old, err := os.ReadFile("testdata/old_task_groups_smith.toml")
	if err != nil {
		t.Fatal(err)
	}
	fixed, err := os.ReadFile("testdata/old_task_groups_smith.fixed.toml")
	if err != nil {
		t.Fatal(err)
	}
	old = []byte(lf(old))
	if _, err := workspace.Parse(old); err == nil {
		t.Fatal("the old file must not load")
	}
	f, ok := RemovedKeysFinding(old)
	if !ok || f.OK || f.Detail != "smith.toml uses [[task_groups]], which is now [[groups]]: run smith doctor --fix" {
		t.Fatalf("finding = %+v", f)
	}
	if f.Fix == nil || !f.Fix.Safe || len(f.Fix.Files) != 1 || f.Fix.Files[0].Path != workspace.FileName {
		t.Fatalf("fix = %+v", f.Fix)
	}
	if got := f.Fix.Files[0].Content; got != lf(fixed) {
		t.Fatalf("fix result:\n%s\nwant:\n%s", got, lf(fixed))
	}
	if cfg, err := workspace.Parse([]byte(f.Fix.Files[0].Content)); err != nil || len(cfg.Groups) != 1 {
		t.Fatalf("the fixed file must load: %v", err)
	}
	if _, ok := RemovedKeysFinding([]byte(f.Fix.Files[0].Content)); ok {
		t.Error("a fixed file has nothing left to report")
	}
}
