package upgrade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/declarationspecifiers"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func convertedWorkspace(t *testing.T, rootScript string) (*session.Session, string) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		file := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("packages/kit/package.json", `{"scripts": {"build": "tsc && node ../../declaration-specifiers.mjs dist/src"}}`)
	write("packages/plain/package.json", `{"scripts": {"build": "tsc"}}`)
	if rootScript != "" {
		write(declarationspecifiers.LegacyFile, rootScript)
	}
	graph := &projectgraph.Graph{Projects: []*projectgraph.Project{
		{Project: protocol.Project{Name: "kit", Root: "packages/kit"}, Plugin: "node"},
		{Project: protocol.Project{Name: "plain", Root: "packages/plain"}, Plugin: "node"},
	}}
	return &session.Session{Root: dir, Graph: graph}, dir
}

func TestUpgradeMovesBuildsOffTheRootScriptAndDeletesIt(t *testing.T) {
	s, dir := convertedWorkspace(t, declarationspecifiers.LegacyContent)
	var r Result
	if err := migrateDeclarationScripts(s, false, &r); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "packages/kit/package.json"))
	if !strings.Contains(string(raw), "tsc && smith declaration-specifiers dist/src") {
		t.Fatalf("%s", raw)
	}
	if _, err := os.Stat(filepath.Join(dir, declarationspecifiers.LegacyFile)); err == nil {
		t.Fatal("the root script was left")
	}
	if len(r.Files) != 2 || r.Files[0] != "packages/kit/package.json" || r.Files[1] != declarationspecifiers.LegacyFile {
		t.Fatalf("files %v", r.Files)
	}
}

func TestUpgradeCheckChangesNothing(t *testing.T) {
	s, dir := convertedWorkspace(t, declarationspecifiers.LegacyContent)
	var r Result
	if err := migrateDeclarationScripts(s, true, &r); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "packages/kit/package.json"))
	if !strings.Contains(string(raw), "node ../../declaration-specifiers.mjs") {
		t.Fatalf("check wrote: %s", raw)
	}
	if _, err := os.Stat(filepath.Join(dir, declarationspecifiers.LegacyFile)); err != nil {
		t.Fatal("check deleted the root script")
	}
	if len(r.Files) == 0 {
		t.Fatal("check reported nothing")
	}
}

func TestUpgradeKeepsAnEditedRootScript(t *testing.T) {
	s, dir := convertedWorkspace(t, declarationspecifiers.LegacyContent+"// mine\n")
	var r Result
	if err := migrateDeclarationScripts(s, false, &r); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, declarationspecifiers.LegacyFile)); err != nil {
		t.Fatal("an edited script was deleted")
	}
	if len(r.Hints) != 1 {
		t.Fatalf("hints %v", r.Hints)
	}
}
