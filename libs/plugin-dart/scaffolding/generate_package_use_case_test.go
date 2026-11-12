package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	var ran []string
	real := Runner
	defer func() { Runner = real }()
	Runner = func(dir string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		if argv[1] == "create" {
			// What flutter create writes, as far as smith cares.
			app := filepath.Join(dir, filepath.FromSlash(argv[len(argv)-1]))
			for _, f := range []string{"web/index.html", "pubspec.yaml", "pubspec.lock", "lib/main.dart", "acme_corp_web.iml", ".gitignore"} {
				_ = os.MkdirAll(filepath.Dir(filepath.Join(app, f)), 0o755)
				_ = os.WriteFile(filepath.Join(app, f), []byte("x"), 0o644)
			}
		}
		return nil
	}
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Workspace: info, Stack: FlutterStack})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "tool", Directory: "apps/tool", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ran, "|") != "dart pub get|flutter create --no-pub --empty --platforms web --project-name acme_corp_web -t app apps/web|flutter pub get|flutter pub get" {
		t.Fatalf("ran %q", ran)
	}
	files := strings.Join(res.Files, ",")
	if !strings.Contains(files, "apps/web/web/index.html") || strings.Contains(files, "pubspec.lock") || strings.Contains(files, ".iml") {
		t.Fatalf("files %s", files)
	}
	if ignore, _ := os.ReadFile(filepath.Join(ws, ".gitignore")); strings.Count(string(ignore), ".dart_tool/") != 1 {
		t.Fatalf(".dart_tool/ is ignored once:\n%s", ignore)
	}
	root, _ := os.ReadFile(filepath.Join(ws, "pubspec.yaml"))
	if !strings.Contains(string(root), "name: acme_corp_workspace\n") || !strings.Contains(string(root), "workspace:\n  - packages/core\n  - apps/web\n  - apps/tool\n") {
		t.Fatalf("root pubspec:\n%s", root)
	}
	found, err := discovery.Discover(ws, []string{"pubspec.yaml", "packages/core/pubspec.yaml", "apps/web/pubspec.yaml", "apps/tool/pubspec.yaml"})
	if err != nil || len(found.Projects) != 3 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	core, web, tool := found.Projects[0], found.Projects[1], found.Projects[2]
	if core.Kind != protocol.ProjectPackage || core.Exports[0] != "acme_corp_core" || core.Tasks["test"].Command[0] != "dart" {
		t.Errorf("core %+v", core)
	}
	if web.Kind != protocol.ProjectApp || web.Stack != FlutterStack || web.Tasks["build"].Command[0] != "flutter" {
		t.Errorf("web %+v", web)
	}
	if tool.Tasks["build"].Command[1] != "build" {
		t.Errorf("tool %+v", tool)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui", Workspace: info, Stack: FlutterStack}); err == nil {
		t.Fatal("the flutter stack makes apps only")
	}
}

func TestMembers(t *testing.T) {
	src := "name: w\nworkspace:\n  - apps/a\n  - 'libs/b'\n\nenvironment:\n  sdk: ^3.6.0\n"
	out, err := AddMember(src, "libs/c")
	if err != nil || out != "name: w\nworkspace:\n  - apps/a\n  - 'libs/b'\n  - libs/c\n\nenvironment:\n  sdk: ^3.6.0\n" {
		t.Fatalf("%v %q", err, out)
	}
	if out, ok := DropMember(out, "libs/b"); !ok || out != "name: w\nworkspace:\n  - apps/a\n  - libs/c\n\nenvironment:\n  sdk: ^3.6.0\n" {
		t.Fatalf("%q", out)
	}
	if _, err := AddMember("name: single\n", "x"); err == nil {
		t.Fatal("a pubspec without a workspace is refused")
	}
}
