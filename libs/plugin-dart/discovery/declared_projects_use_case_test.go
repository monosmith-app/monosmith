package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declaredWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for name, body := range files {
		full := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

const toolsPubspec = "name: acme_tools\nversion: 0.1.0\ndependencies:\n  acme_core: any\n"

func TestDeclaredFoldersOfAPackageAreProjectsOfThatPackage(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"pubspec.yaml":                     "name: ws\nworkspace:\n  - packages/tools\n",
		"packages/tools/pubspec.yaml":      toolsPubspec,
		"packages/tools/bin/main.dart":     "void main() {}\n",
		"packages/tools/lib/x/test/a.dart": "",
		"packages/tools/lib/x/x.dart":      "",
	})
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "cli", Root: "packages/tools/bin"}, {Name: "x", Root: "packages/tools/lib/x"}})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("%v %v", res.Projects, err)
	}
	app, lib := res.Projects[0], res.Projects[1]
	if app.Kind != protocol.ProjectApp || lib.Kind != protocol.ProjectLib || app.Language != "dart" || len(app.Exports) != 0 {
		t.Errorf("an app for the folder with a main.dart: %+v %+v", app, lib)
	}
	if !slices.Equal(app.Imports, []string{"acme_core"}) || !slices.Contains(app.Tasks["lint"].Inputs, "{workspaceRoot}/packages/tools/pubspec.yaml") {
		t.Errorf("imports and inputs: %+v", app)
	}
	if got := strings.Join(app.Tasks["start"].Command, " "); got != "dart run main.dart" || app.Tasks["build"].Command == nil {
		t.Errorf("start: %q", got)
	}
	if lib.Tasks["test"].Command == nil || lib.Tasks["start"].Command != nil || lib.Tasks["build"].Command != nil {
		t.Errorf("a lib has test when it has a test folder, and neither build nor start: %+v", lib.Tasks)
	}
}

func TestADeclaredDartFolderIsRefusedWhenItIsNotAFolderOfAPackage(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"pubspec.yaml":                "name: ws\nworkspace:\n  - packages/tools\n",
		"packages/tools/pubspec.yaml": toolsPubspec,
		"packages/tools/lib/a.dart":   "",
		"packages/own/pubspec.yaml":   "name: own\n",
		"packages/ui/pubspec.yaml":    "name: ui\ndependencies:\n  flutter:\n    sdk: flutter\n",
		"packages/ui/lib/ui.dart":     "",
		"loose/a.dart":                "",
	})
	cases := map[string]struct {
		root, want string
		kind       protocol.ProjectKind
	}{
		"missing":     {"packages/tools/none", "does not exist", ""},
		"own":         {"packages/own", "pubspec.yaml of its own", ""},
		"no manifest": {"loose", "no pubspec.yaml of a package", ""},
		"flutter":     {"packages/ui/lib", "Flutter package", ""},
		"package":     {"packages/tools/lib", "cannot be released on its own", protocol.ProjectPackage},
		"app":         {"packages/tools/lib", "runs main.dart", protocol.ProjectApp},
	}
	for name, c := range cases {
		_, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "x", Root: c.root, Kind: c.kind}})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
