package workspacejoin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projects(roots ...string) []protocol.Project {
	out := make([]protocol.Project, len(roots))
	for i, r := range roots {
		out[i] = protocol.Project{Name: filepath.Base(r), Root: r}
	}
	return out
}

func writeRoot(t *testing.T, text string) string {
	t.Helper()
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "package.json"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws
}

func TestJoinSaysWhichProjectsTheWorkspacesCover(t *testing.T) {
	ws := writeRoot(t, "{\n  \"name\": \"w\",\n  \"workspaces\": [\"packages/*\"],\n  \"scripts\": {\"x\": \"a && b\"}\n}\n")
	res, err := Join(ws, protocol.JoinParams{Projects: projects(".", "packages/lib", "apps/api", "apps/web", "tools/x")})
	if err != nil {
		t.Fatal(err)
	}
	joined := map[string]bool{}
	for _, s := range res.Projects {
		joined[s.Project] = s.Joined
	}
	if len(res.Projects) != 4 || !joined["lib"] || joined["api"] || joined["web"] || joined["x"] {
		t.Fatalf("states %+v", res.Projects)
	}
	fix := res.Projects[1].Fix
	if fix == nil || !fix.Safe || !reflect.DeepEqual(fix.AppliesTo, []string{"api", "web", "x"}) || len(fix.Files) != 1 || fix.Steps[0].Argv[1] != "install" {
		t.Fatalf("fix %+v", fix)
	}
	if res.Projects[2].Fix != fix || res.Projects[0].Fix != nil {
		t.Error("the unjoined projects must share one fix, and a joined one has none")
	}
	content := fix.Files[0].Content
	var pkg struct{ Workspaces []string }
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pkg.Workspaces, []string{"packages/*", "apps/*", "tools/*"}) || !strings.Contains(content, `"a && b"`) || !strings.HasPrefix(content, "{\n  \"name\": \"w\",") {
		t.Errorf("the manifest must keep everything else as written:\n%s", content)
	}
	if !strings.Contains(fix.Description, `"apps/*", "tools/*"`) {
		t.Errorf("description %q", fix.Description)
	}
}

func TestJoinWritesARootManifestWhenThereIsNone(t *testing.T) {
	ws := t.TempDir()
	res, err := Join(ws, protocol.JoinParams{Projects: projects("apps/api")})
	if err != nil || len(res.Projects) != 1 || res.Projects[0].Joined || res.Projects[0].Fix == nil {
		t.Fatalf("%+v, %v", res, err)
	}
	var pkg struct {
		Private    bool
		Workspaces []string
	}
	if err := json.Unmarshal([]byte(res.Projects[0].Fix.Files[0].Content), &pkg); err != nil || !pkg.Private || !reflect.DeepEqual(pkg.Workspaces, []string{"apps/*"}) {
		t.Fatalf("manifest %+v, %v", pkg, err)
	}
}

func TestJoinLeavesTheObjectFormToThePerson(t *testing.T) {
	ws := writeRoot(t, `{"workspaces":{"packages":["libs/*"]}}`)
	res, err := Join(ws, protocol.JoinParams{Projects: projects("libs/a", "apps/b")})
	if err != nil || !res.Projects[0].Joined || res.Projects[1].Joined || res.Projects[1].Fix != nil || res.Projects[1].Problem == "" {
		t.Fatalf("%+v, %v", res, err)
	}
}

func TestUnjoinedNamesWhatNpmInstallDoesNotReach(t *testing.T) {
	ws := writeRoot(t, `{"workspaces":["./apps/*","libs/core/"]}`)
	got := Unjoined(ws, projects(".", "apps/web", "libs/core", "libs/other"))
	if !reflect.DeepEqual(got, []string{"other"}) {
		t.Fatalf("got %v", got)
	}
}

func TestPatternIsTheParentGlobOrTheFolder(t *testing.T) {
	if Pattern("apps/web") != "apps/*" || Pattern("tool") != "tool" || Pattern("a/b/c") != "a/b/*" {
		t.Error(Pattern("apps/web"), Pattern("tool"), Pattern("a/b/c"))
	}
}
