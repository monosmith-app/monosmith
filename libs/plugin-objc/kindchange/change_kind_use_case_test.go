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
	_ = os.WriteFile(filepath.Join(dir, "X.podspec"), []byte("# smithKind: lib\nPod::Spec.new\n"), 0o644)
	project := protocol.Project{Name: "x", Root: "libs/x", Language: "objc"}

	res, err := Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectPackage, NewRoot: "packages/x"})
	if err != nil || len(res.Files) != 1 || res.Files[0].Path != "libs/x/X.podspec" || res.Files[0].Content != "# smithKind: package\nPod::Spec.new\n" {
		t.Fatalf("%+v %v", res, err)
	}

	_ = os.WriteFile(filepath.Join(dir, "X.podspec"), []byte("Pod::Spec.new\n"), 0o644)
	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectLib})
	if err != nil || res.Files[0].Content != "# smithKind: lib\nPod::Spec.new\n" {
		t.Fatalf("a missing comment is added on top: %+v %v", res, err)
	}

	res, err = Change(ws, protocol.ChangeKindParams{Project: project, To: protocol.ProjectApp})
	if err != nil || res.Refused == "" || len(res.Files) != 0 {
		t.Errorf("an app is refused: %+v %v", res, err)
	}
	if _, err := Change(t.TempDir(), protocol.ChangeKindParams{Project: project, To: protocol.ProjectLib}); err == nil {
		t.Error("a project without a podspec is an error")
	}
}
