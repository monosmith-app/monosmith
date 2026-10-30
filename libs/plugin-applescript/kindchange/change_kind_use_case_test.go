package kindchange

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestChange(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs", "x")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "main.applescript"), []byte("-- smithKind: lib\non run\nend run\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "other.applescript"), []byte(""), 0o644)
	project := protocol.Project{Name: "x", Root: "libs/x", Language: "applescript"}

	res, err := Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp, NewRoot: "apps/x"})
	if err != nil || len(res.Files) != 1 || res.Files[0].Path != "libs/x/main.applescript" || res.Files[0].Content != "-- smithKind: app\non run\nend run\n" {
		t.Fatalf("%+v %v", res, err)
	}

	_ = os.WriteFile(filepath.Join(dir, "main.applescript"), []byte("on run\nend run\n"), 0o644)
	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectLib})
	if err != nil || res.Files[0].Content != "-- smithKind: lib\non run\nend run\n" {
		t.Fatalf("a missing comment is added on top: %+v %v", res, err)
	}

	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectPackage})
	if err != nil || res.Refused == "" || len(res.Files) != 0 {
		t.Errorf("a package is refused: %+v %v", res, err)
	}
	if _, err := Change(t.TempDir(), protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp}); err == nil {
		t.Error("a project without a script is an error")
	}
}
