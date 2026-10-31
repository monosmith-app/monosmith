package scaffolding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func testRenderer() FileRenderer {
	return FileRenderer{
		ID: "test.tsconfig", Version: 1, File: "tsconfig.json",
		Templates: func(protocol.GenerateParams) (TemplateSet, bool) {
			return TemplateSet{Slice: map[string]string{"[[.Dir]]/tsconfig.json": "{ \"extends\": \"[[.Up]]base.json\" }\n"}}, true
		},
	}
}

func deriveParams(owned ...string) protocol.DeriveParams {
	return protocol.DeriveParams{
		Projects: []protocol.Project{{Name: "a", Root: "libs/a", Kind: protocol.ProjectLib}},
		Owned:    owned,
	}
}

func TestDeriveRendersTheTemplateOfAnOwnedFileThatIsGone(t *testing.T) {
	res, err := testRenderer().Derive(t.TempDir(), deriveParams("libs/a/tsconfig.json"))
	if err != nil || len(res.Files) != 1 {
		t.Fatal(err, res.Files)
	}
	f := res.Files[0]
	if f.Path != "libs/a/tsconfig.json" || f.Mode != protocol.DeriveFile || f.Content != "{ \"extends\": \"../../base.json\" }\n" {
		t.Errorf("the whole file comes back as the template renders it: %+v", f)
	}
}

func TestDeriveTakesAFileThatIsWhatTheTemplateRenders(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/a/tsconfig.json", "{ \"extends\": \"../../base.json\" }\r\n")
	res, _ := testRenderer().Derive(ws, deriveParams())
	if len(res.Files) != 1 {
		t.Fatalf("a file as scaffolded is derived, whatever its line endings: %+v", res.Files)
	}
}

func TestDeriveLeavesAPersonsFileAlone(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/a/tsconfig.json", "{ \"extends\": \"./mine.json\" }\n")
	if res, _ := testRenderer().Derive(ws, deriveParams()); len(res.Files) != 0 {
		t.Errorf("a file that differs and is not owned is the person's: %+v", res.Files)
	}
	if res, _ := testRenderer().Derive(ws, deriveParams("libs/a/tsconfig.json")); len(res.Files) != 1 {
		t.Errorf("an owned file is derived, so its drift is reported: %+v", res.Files)
	}
}

func TestDeriveSkipsAProjectThePluginDidNotScaffold(t *testing.T) {
	r := testRenderer()
	r.Templates = func(protocol.GenerateParams) (TemplateSet, bool) { return TemplateSet{}, false }
	if res, _ := r.Derive(t.TempDir(), deriveParams("libs/a/tsconfig.json")); len(res.Files) != 0 {
		t.Errorf("no template, no file: %+v", res.Files)
	}
}

func TestStackChoicesReadsTheSavedSelection(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "smith.toml", "[stacks.mine]\nchoices = [\"react\", \"vitest\"]\n")
	if got := StackChoices(ws, "mine"); len(got) != 2 || got[0] != "react" {
		t.Errorf("got %v", got)
	}
	if got := StackChoices(ws, "other"); got != nil {
		t.Errorf("an unknown stack has none: %v", got)
	}
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
