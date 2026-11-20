package workspacejoin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestJoinReportsNothingWithoutAUvWorkspace(t *testing.T) {
	ws := t.TempDir()
	res, err := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "a", Root: "apps/a"}}})
	if err != nil || len(res.Projects) != 0 {
		t.Fatalf("%+v, %v", res, err)
	}
	if err := os.WriteFile(filepath.Join(ws, "pyproject.toml"), []byte("[tool.ruff]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if res, _ := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "a", Root: "apps/a"}}}); len(res.Projects) != 0 {
		t.Fatalf("a pyproject.toml without [tool.uv.workspace] is no uv workspace: %+v", res)
	}
}

func TestJoinAddsTheMembersTheUvWorkspaceLacks(t *testing.T) {
	ws := t.TempDir()
	src := "[project]\nname = \"ws\"\n\n[tool.uv.workspace]\nmembers = [\"libs/*\"]\n"
	if err := os.WriteFile(filepath.Join(ws, "pyproject.toml"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "core", Root: "libs/core"}, {Name: "api", Root: "apps/api"}, {Name: "web", Root: "apps/web"}}})
	if err != nil || len(res.Projects) != 3 || !res.Projects[0].Joined || res.Projects[1].Joined {
		t.Fatalf("%+v, %v", res, err)
	}
	fix := res.Projects[1].Fix
	if fix == nil || res.Projects[2].Fix != fix || len(fix.AppliesTo) != 2 {
		t.Fatalf("fix %+v", fix)
	}
	got := fix.Files[0].Content
	if !strings.Contains(got, `members = ["libs/*", "apps/*"]`) || !strings.HasPrefix(got, "[project]\nname = \"ws\"") {
		t.Errorf("content:\n%s", got)
	}
}

func TestCoveredMatchesAGlobOrTheFolder(t *testing.T) {
	if !Covered([]string{"./apps/*"}, "apps/x") || !Covered([]string{"tools/cli/"}, "tools/cli") || Covered([]string{"apps/*"}, "libs/x") {
		t.Error("coverage")
	}
}
