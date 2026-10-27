package releaseassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompanionsPreferTheProjectsFilesOverTheWorkspaces(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	for _, f := range []string{"README.md", "LICENSE-MIT", "LICENSE-APACHE", "apps/cli/readme.md", "NOTICE"} {
		_ = os.WriteFile(filepath.Join(ws, filepath.FromSlash(f)), []byte(f), 0o644)
	}
	var got []string
	for _, e := range Companions(ws, "apps/cli") {
		got = append(got, e.Name)
	}
	if strings.Join(got, ",") != "LICENSE-APACHE,LICENSE-MIT,readme.md" {
		t.Errorf("%v", got)
	}
	if len(Companions(t.TempDir(), "apps/none")) != 0 {
		t.Error("nothing to add when there is nothing")
	}
}
