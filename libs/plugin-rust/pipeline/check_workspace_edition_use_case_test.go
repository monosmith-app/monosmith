package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckWorkspaceEditionOffersToShareTheEdition(t *testing.T) {
	ws := t.TempDir()
	for dir, body := range map[string]string{
		"libs/a": "[package]\nname = \"a\"\nedition = \"2021\"\n",
		"libs/b": "[package]\nname = \"b\"\nedition.workspace = true\n",
	} {
		if err := os.MkdirAll(filepath.Join(ws, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ws, filepath.FromSlash(dir), "Cargo.toml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	projects := []protocol.Project{{Name: "a", Root: "libs/a"}, {Name: "b", Root: "libs/b"}}
	f, ok := CheckWorkspaceEdition(ws, "[workspace]\nmembers = [\"libs/a\", \"libs/b\"]\n", projects)
	if !ok || !f.Warning || f.Fix == nil || f.Fix.Safe || len(f.Fix.Files) != 2 {
		t.Fatalf("want an unsafe fix editing the root and one crate, got %+v", f)
	}
	if !strings.Contains(f.Fix.Files[0].Content, "[workspace.package]\nedition = \"2021\"") || f.Fix.Files[1].Path != "libs/a/Cargo.toml" || !strings.Contains(f.Fix.Files[1].Content, "edition.workspace = true") {
		t.Fatalf("files %+v", f.Fix.Files)
	}
	if _, ok := CheckWorkspaceEdition(ws, "[workspace]\n[workspace.package]\nedition = \"2021\"\n", projects[1:]); ok {
		t.Fatal("a crate that inherits needs no finding")
	}
}
