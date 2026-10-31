package deriving

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReferencesFollowTheGraphAndSkipProjectsWithoutAConfig(t *testing.T) {
	root := t.TempDir()
	put(t, root, "libs/core/tsconfig.lib.json", "{}")
	put(t, root, "libs/util/tsconfig.json", "{}")
	put(t, root, "libs/bare/package.json", "{}")
	put(t, root, "apps/web/tsconfig.lib.json", `{"extends": "../../tsconfig.json"}`)
	params := protocol.DeriveParams{
		Projects: []protocol.Project{
			{Name: "web", Root: "apps/web"}, {Name: "core", Root: "libs/core"}, {Name: "util", Root: "libs/util"}, {Name: "bare", Root: "libs/bare"},
		},
		Graph: map[string][]string{"web": {"bare", "core", "util", "gone"}},
	}
	res, err := Derive(root, params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 1 {
		t.Fatalf("files: %+v", res.Files)
	}
	f := res.Files[0]
	refs := f.Keys["references"].([]any)
	if f.Path != "apps/web/tsconfig.lib.json" || f.Mode != protocol.DeriveKeys || len(refs) != 2 ||
		refs[0].(map[string]any)["path"] != "../../libs/core/tsconfig.lib.json" || refs[1].(map[string]any)["path"] != "../../libs/util/tsconfig.json" {
		t.Fatalf("derived %+v", f)
	}
}

func TestAnOwnedFileThatIsGoneIsProposedAgain(t *testing.T) {
	root := t.TempDir()
	params := protocol.DeriveParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}, Owned: []string{"apps/web/tsconfig.lib.json"}}
	res, _ := Derive(root, params)
	if len(res.Files) != 1 || len(res.Files[0].Keys["references"].([]any)) != 0 {
		t.Fatalf("%+v", res.Files)
	}
}

func TestJestDisplayNameIsARegionBetweenMarkers(t *testing.T) {
	root := t.TempDir()
	put(t, root, "libs/a/jest.config.ts", "export default {\n  // smith:begin node.jest.display-name\n  displayName: 'x',\n  // smith:end node.jest.display-name\n}\n")
	put(t, root, "libs/b/jest.config.ts", "export default {}\n")
	res, _ := Derive(root, protocol.DeriveParams{Projects: []protocol.Project{{Name: "a", Root: "libs/a", Exports: []string{"@s/a"}}, {Name: "b", Root: "libs/b"}}})
	if len(res.Files) != 1 || res.Files[0].Mode != protocol.DeriveRegion || res.Files[0].Content != "  displayName: '@s/a',\n" {
		t.Fatalf("%+v", res.Files)
	}
}
