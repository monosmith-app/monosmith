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

const toolsGemspec = "Gem::Specification.new do |s|\n  s.name = 'acme_tools'\n  s.version = '0.1.0'\n  s.executables = ['gen']\n  s.add_dependency 'acme_core'\nend\n"

func TestDeclaredFoldersOfAGemAreProjectsOfThatGem(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"gems/tools/acme_tools.gemspec":   toolsGemspec,
		"gems/tools/exe/gen":              "puts 1\n",
		"gems/tools/lib/x/spec/x_spec.rb": "",
		"gems/tools/lib/x/y.rb":           "",
	})
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{
		{Name: "gen", Root: "gems/tools/exe", Kind: protocol.ProjectApp},
		{Name: "x", Root: "gems/tools/lib/x"},
	})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("%v %v", res.Projects, err)
	}
	app, lib := res.Projects[0], res.Projects[1]
	if got := strings.Join(app.Tasks["start"].Command, " "); got != "bundle exec ruby gen" || app.Language != "ruby" || len(app.Exports) != 0 {
		t.Errorf("an app runs the executable that is in its folder: %q %+v", got, app)
	}
	if !slices.Equal(app.Imports, []string{"acme_core"}) || !slices.Contains(app.Tasks["lint"].Inputs, "{workspaceRoot}/gems/tools/acme_tools.gemspec") {
		t.Errorf("imports and inputs: %+v", app)
	}
	if lib.Kind != protocol.ProjectLib || lib.Tasks["test"].Command == nil || lib.Tasks["start"].Command != nil {
		t.Errorf("a lib with specs has test and no start: %+v", lib)
	}
}

func TestADeclaredRubyFolderIsRefusedWhenItIsNotAFolderOfAGem(t *testing.T) {
	ws := declaredWorkspace(t, map[string]string{
		"gems/tools/acme_tools.gemspec": toolsGemspec,
		"gems/tools/lib/a.rb":           "",
		"gems/own/own.gemspec":          "s.name = 'own'\n",
		"loose/a.rb":                    "",
	})
	cases := map[string]struct {
		root, want string
		kind       protocol.ProjectKind
	}{
		"missing":     {"gems/tools/none", "does not exist", ""},
		"own":         {"gems/own", "gemspec of its own", ""},
		"no manifest": {"loose", "no gemspec", ""},
		"package":     {"gems/tools/lib", "cannot be released on its own", protocol.ProjectPackage},
		"app":         {"gems/tools/lib", "none is in the folder", protocol.ProjectApp},
	}
	for name, c := range cases {
		_, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "x", Root: c.root, Kind: c.kind}})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
