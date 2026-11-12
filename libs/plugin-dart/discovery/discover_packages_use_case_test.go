package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("pubspec.yaml", "name: acme_workspace\npublish_to: none\nworkspace:\n  - apps/web\n  - packages/core\n")
	write("apps/web/pubspec.yaml", `name: acme_web
version: 1.2.0
resolution: workspace
dependencies:
  flutter:
    sdk: flutter
  acme_core: ^0.0.0
dev_dependencies:
  flutter_test:
    sdk: flutter
smith:
  kind: app
  stack: flutter
`)
	write("apps/web/test/widget_test.dart", "")
	write("packages/core/pubspec.yaml", "name: acme_core\nversion: 0.0.0\ndev_dependencies:\n  test: ^1.25.0\n")
	write("apps/tool/pubspec.yaml", "name: acme_tool\nversion: 0.0.0\nsmith:\n  kind: app\n")
	res, err := Discover(ws, []string{"pubspec.yaml", "apps/web/pubspec.yaml", "packages/core/pubspec.yaml", "apps/tool/pubspec.yaml"})
	if err != nil || len(res.Projects) != 3 {
		t.Fatalf("the workspace root is not a project: %v %+v", err, res.Projects)
	}
	web, core, tool := res.Projects[0], res.Projects[1], res.Projects[2]
	if web.Kind != protocol.ProjectApp || web.Stack != "flutter" || web.Version != "1.2.0" || web.Exports[0] != "acme_web" || web.PackageName != "acme_web" || web.Imports[0] != "acme_core" || web.Imports[1] != "flutter_test" {
		t.Errorf("web %+v", web)
	}
	if build := web.Tasks["build"]; build.Command[0] != "flutter" || build.Command[len(build.Command)-1] != "{workspaceRoot}/dist/apps/web" || build.Outputs[0] != "../../dist/apps/web" {
		t.Errorf("web build %+v", build)
	}
	if web.Tasks["test"].Command[0] != "flutter" || web.Tasks["lint"].Command[0] != "flutter" {
		t.Errorf("web tasks %+v", web.Tasks)
	}
	if _, ok := core.Tasks["test"]; ok || core.Kind != protocol.ProjectPackage || core.Tasks["lint"].Command[0] != "dart" {
		t.Errorf("core has no test/ %+v", core)
	}
	if build := tool.Tasks["build"]; build.Command[2] != "cli" || tool.Tasks["start"].Command[1] != "run" {
		t.Errorf("tool %+v", tool.Tasks)
	}
}
