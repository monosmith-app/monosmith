package fileedits

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRetargetDependentsRewritesOnlySameLanguage(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, text string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("apps/web/m.toml", `core = { path = "../../libs/core" }`)
	write("apps/other/m.toml", `core = { path = "../../libs/core" }`)
	edits := New(dir)
	params := protocol.ChangeKindParams{
		Project:  protocol.Project{Root: "libs/core"},
		NewRoot:  "packages/core",
		Projects: []protocol.Project{{Root: "apps/web", Language: "x"}, {Root: "apps/other", Language: "y"}, {Root: "libs/core", Language: "x"}},
	}
	RetargetDependents(edits, params, "x", func(root string) []string { return []string{root + "/m.toml", root + "/missing.toml"} })
	if got, _ := edits.Read("apps/web/m.toml"); got != `core = { path = "../../packages/core" }` {
		t.Fatalf("web = %s", got)
	}
	if got, _ := edits.Read("apps/other/m.toml"); got != `core = { path = "../../libs/core" }` {
		t.Fatalf("other changed: %s", got)
	}
}
