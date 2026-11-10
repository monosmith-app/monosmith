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

const toolsComposer = `{"name": "acme/tools", "bin": ["bin/gen"], "require": {"php": ">=8.2", "acme/core": "*"}}`

func TestDeclaredFoldersOfAPackageAreProjectsOfThatPackage(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"composer.json":                           `{"name": "acme/mono"}`,
		"packages/tools/composer.json":            toolsComposer,
		"packages/tools/bin/gen":                  "<?php\n",
		"packages/tools/src/Util/A.php":           "<?php\n",
		"packages/tools/src/Util/tests/ATest.php": "<?php\n",
	})
	res, err := DiscoverWithDeclared(ws, []string{"composer.json", "packages/tools/composer.json"}, []protocol.DeclaredProject{
		{Name: "gen", Root: "packages/tools/bin", Kind: protocol.ProjectApp},
		{Name: "util", Root: "packages/tools/src/Util"},
	})
	if err != nil || len(res.Projects) != 3 {
		t.Fatalf("the package itself and the two declared folders: %v %v", res.Projects, err)
	}
	app, lib := res.Projects[1], res.Projects[2]
	if got := strings.Join(app.Tasks["start"].Command, " "); got != "php gen" || app.Language != "php" || len(app.Exports) != 0 {
		t.Errorf("an app runs the executable that is in its folder: %q %+v", got, app)
	}
	if !slices.Equal(app.Imports, []string{"acme/core"}) || !slices.Contains(app.Tasks["lint"].Inputs, "{workspaceRoot}/packages/tools/composer.json") {
		t.Errorf("imports and inputs: %+v", app)
	}
	if !slices.Contains(lib.Tasks["lint"].Command, ".") || slices.Contains(lib.Tasks["lint"].Command, "src") {
		t.Errorf("PHPStan analyses the folder whole: %v", lib.Tasks["lint"].Command)
	}
	if lib.Kind != protocol.ProjectLib || lib.Tasks["start"].Command != nil {
		t.Errorf("a lib has no start: %+v", lib)
	}
}

func TestADeclaredPHPFolderIsRefusedWhenItIsNotAFolderOfAPackage(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"composer.json":                `{"name": "acme/mono"}`,
		"packages/tools/composer.json": toolsComposer,
		"packages/tools/src/A.php":     "<?php\n",
		"packages/own/composer.json":   `{"name": "acme/own"}`,
		"loose/a.php":                  "<?php\n",
	})
	cases := map[string]struct {
		root, want string
		kind       protocol.ProjectKind
	}{
		"missing":     {"packages/tools/none", "does not exist", ""},
		"own":         {"packages/own", "composer.json of its own", ""},
		"no manifest": {"loose", "no composer.json in a folder above it", ""},
		"package":     {"packages/tools/src", "cannot be released on its own", protocol.ProjectPackage},
		"app":         {"packages/tools/src", "none is in the folder", protocol.ProjectApp},
	}
	for name, c := range cases {
		_, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "x", Root: c.root, Kind: c.kind}})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
