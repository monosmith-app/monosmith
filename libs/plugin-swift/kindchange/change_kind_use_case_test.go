package kindchange

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestChange(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("libs/core/Package.swift", "// swift-tools-version: 5.9\n// smithKind: lib\nlet package = Package(dependencies: [.package(path: \"../util\"), .package(path: \"Vendored\")])\n")
	write("libs/util/Package.swift", "// swift-tools-version: 5.9\n")
	write("apps/cli/Package.swift", "// swift-tools-version: 5.9\nlet package = Package(dependencies: [.package(path: \"../../libs/core\"), .package(name: \"u\", path: \"../../libs/util\")])\n")
	projects := []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "swift"},
		{Name: "util", Root: "libs/util", Language: "swift"},
		{Name: "cli", Root: "apps/cli", Language: "swift"},
	}
	res, err := Change(ws, protocol.ChangeKindParams{Project: projects[0], Projects: projects, To: protocol.ProjectPackage, NewRoot: "packages/core"})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range res.Files {
		got[f.Path] = f.Content
	}
	if len(got) != 2 {
		t.Fatalf("the project and its dependent change, not the unrelated one: %v", got)
	}
	if want := "// swift-tools-version: 5.9\n// smithKind: package\nlet package = Package(dependencies: [.package(path: \"../../libs/util\"), .package(path: \"Vendored\")])\n"; got["libs/core/Package.swift"] != want {
		t.Errorf("own manifest: %q", got["libs/core/Package.swift"])
	}
	if want := "// swift-tools-version: 5.9\nlet package = Package(dependencies: [.package(path: \"../../packages/core\"), .package(name: \"u\", path: \"../../libs/util\")])\n"; got["apps/cli/Package.swift"] != want {
		t.Errorf("dependent: %q", got["apps/cli/Package.swift"])
	}

	res, err = Change(ws, protocol.ChangeKindParams{Project: projects[1], Projects: projects, To: protocol.ProjectApp, NewRoot: "libs/util"})
	if err != nil || len(res.Files) != 1 || res.Files[0].Content != "// swift-tools-version: 5.9\n// smithKind: app\n" {
		t.Errorf("without a move only the comment changes: %+v %v", res, err)
	}
	if _, err := Change(t.TempDir(), protocol.ChangeKindParams{Project: projects[0], To: protocol.ProjectApp}); err == nil {
		t.Error("a project without a manifest is an error")
	}
}
