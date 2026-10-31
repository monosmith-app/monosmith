package deriving

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func webWithDeps(t *testing.T, file string) protocol.DerivedFile {
	t.Helper()
	root := t.TempDir()
	put(t, root, "libs/a/tsconfig.lib.json", "{}")
	put(t, root, "libs/b/tsconfig.lib.json", "{}")
	put(t, root, "libs/c/tsconfig.lib.json", "{}")
	put(t, root, "apps/web/tsconfig.lib.json", file)
	res, err := Derive(root, protocol.DeriveParams{
		Projects: []protocol.Project{{Name: "web", Root: "apps/web"}, {Name: "a", Root: "libs/a"}, {Name: "b", Root: "libs/b"}, {Name: "c", Root: "libs/c"}},
		Graph:    map[string][]string{"web": {"a", "b", "c"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if f.Path == "apps/web/tsconfig.lib.json" {
			return f
		}
	}
	t.Fatal("web is not derived")
	return protocol.DerivedFile{}
}

func paths(f protocol.DerivedFile) []string {
	var out []string
	for _, r := range f.Keys["references"].([]any) {
		out = append(out, r.(map[string]any)["path"].(string))
	}
	return out
}

func TestReferencesKeepTheOrderOfTheFileWhenTheSetIsTheSame(t *testing.T) {
	f := webWithDeps(t, `{"references": [{"path": "../../libs/c/tsconfig.lib.json"}, {"path": "../../libs/a/tsconfig.lib.json"}, {"path": "../../libs/b/tsconfig.lib.json"}]}`)
	got := paths(f)
	if len(got) != 3 || got[0] != "../../libs/c/tsconfig.lib.json" || got[1] != "../../libs/a/tsconfig.lib.json" || got[2] != "../../libs/b/tsconfig.lib.json" {
		t.Fatalf("order %v", got)
	}
}

func TestReferencesAddNewEntriesLastAndDropOldOnesKeepingExtraFields(t *testing.T) {
	f := webWithDeps(t, `{"references": [{"path": "../../libs/gone/tsconfig.lib.json"}, {"path": "../../libs/b/tsconfig.lib.json", "prepend": true}]}`)
	got := paths(f)
	if len(got) != 3 || got[0] != "../../libs/b/tsconfig.lib.json" || got[1] != "../../libs/a/tsconfig.lib.json" || got[2] != "../../libs/c/tsconfig.lib.json" {
		t.Fatalf("order %v", got)
	}
	if f.Keys["references"].([]any)[0].(map[string]any)["prepend"] != true {
		t.Fatalf("the extra field of a kept entry is lost: %v", f.Keys)
	}
}

func TestReferencesOfAFileWithCommentsKeepTheWantedOrder(t *testing.T) {
	f := webWithDeps(t, "{ // comment\n \"references\": [] }")
	if got := paths(f); len(got) != 3 || got[0] != "../../libs/a/tsconfig.lib.json" {
		t.Fatalf("wanted order for an unreadable file, got %v", got)
	}
}
