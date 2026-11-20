package workspacejoin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func write(t *testing.T, dir, rel, text string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestJoinReportsNothingWithoutASolution(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/api/Api.csproj", "<Project/>")
	res, err := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}})
	if err != nil || len(res.Projects) != 0 {
		t.Fatalf("%+v, %v", res, err)
	}
}

func TestJoinAddsAProjectTheSolutionDoesNotList(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/api/Api.csproj", "<Project/>")
	write(t, ws, "libs/core/Core.csproj", "<Project/>")
	write(t, ws, "Shop.sln", `Project("{FAE04EC0}") = "Api", "apps\\api\\Api.csproj", "{1}"`+"\nEndProject\n")
	res, err := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("%+v, %v", res, err)
	}
	if !res.Projects[0].Joined || res.Projects[1].Joined {
		t.Fatalf("states %+v", res.Projects)
	}
	fix := res.Projects[1].Fix
	if fix == nil || !fix.Safe || len(fix.Steps) != 1 || fix.Steps[0].Argv[2] != "Shop.sln" || fix.Steps[0].Argv[4] != "libs/core/Core.csproj" {
		t.Fatalf("fix %+v", fix)
	}
}

func TestJoinAsksWhichSolutionWhenThereAreSeveral(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/api/Api.csproj", "<Project/>")
	write(t, ws, "A.sln", "")
	write(t, ws, "B.slnx", "<Solution/>")
	res, err := Join(ws, protocol.JoinParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}})
	if err != nil || len(res.Projects) != 1 || res.Projects[0].Fix == nil {
		t.Fatalf("%+v, %v", res, err)
	}
	if c := res.Projects[0].Fix.Choices; len(c) != 1 || len(c[0].Options) != 2 || c[0].Options[1].Label != "B.slnx" {
		t.Fatalf("choices %+v", c)
	}
}
