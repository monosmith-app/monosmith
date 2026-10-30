package kindchange

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func workspaceWith(t *testing.T, rel, body string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestChange(t *testing.T) {
	project := protocol.Project{Name: "x", Root: "libs/x", Language: "lua"}

	ws := workspaceWith(t, "libs/x/x-0.0.0-1.rockspec", "-- smithKind: lib\npackage = \"x\"\n")
	res, err := Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp, NewRoot: "apps/x"})
	if err != nil || len(res.Files) != 1 || res.Files[0].Path != "libs/x/x-0.0.0-1.rockspec" || res.Files[0].Content != "-- smithKind: app\npackage = \"x\"\n" {
		t.Fatalf("%+v %v", res, err)
	}

	ws = workspaceWith(t, "libs/x/x-0.0.0-1.rockspec", "package = \"x\"\n")
	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp})
	if err != nil || res.Files[0].Content != "-- smithKind: app\npackage = \"x\"\n" {
		t.Fatalf("a missing comment is added on top: %+v %v", res, err)
	}

	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectPackage})
	if err != nil || res.Refused == "" || len(res.Files) != 0 {
		t.Errorf("a package is refused: %+v %v", res, err)
	}

	if _, err := Change(t.TempDir(), protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp}); err == nil {
		t.Error("a project without a rockspec is an error")
	}
}
