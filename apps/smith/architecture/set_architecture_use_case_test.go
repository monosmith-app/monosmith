package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const plain = "schema = 1\n\n[workspace]\nname = \"acme\" # stays\n\n[ci]\nproviders = [\"github\"]\n"

func workspaceWith(t *testing.T, toml string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspace.FileName), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func fileOf(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func currentName(t *testing.T, root string) string {
	t.Helper()
	a, err := Current(root)
	if err != nil {
		t.Fatal(err)
	}
	return a.Name
}

func TestAWorkspaceWithoutTheTableUsesNoArchitecture(t *testing.T) {
	root := workspaceWith(t, plain)
	if got := currentName(t, root); got != None {
		t.Fatalf("got %q", got)
	}
}

func TestChoosingAnArchitectureAddsItsTableAndNothingElseChanges(t *testing.T) {
	root := workspaceWith(t, plain)
	if err := Set(root, "vertical-slices"); err != nil {
		t.Fatal(err)
	}
	if got, want := fileOf(t, root), plain+"\n[conventions.vertical-slices]\n"; got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := currentName(t, root); got != "vertical-slices" {
		t.Fatalf("got %q", got)
	}
	// Choosing it again changes nothing.
	before := fileOf(t, root)
	if err := Set(root, "vertical-slices"); err != nil {
		t.Fatal(err)
	}
	if fileOf(t, root) != before {
		t.Fatal("the same choice rewrote the file")
	}
}

func TestChoosingNoneRemovesTheTableAndWhatItHeld(t *testing.T) {
	with := plain + "\n# slices\n[conventions.vertical-slices]\nlanguages = [\"go\"]\n\n[release]\nbranch = \"main\"\n"
	root := workspaceWith(t, with)
	if got := currentName(t, root); got != "vertical-slices" {
		t.Fatalf("got %q", got)
	}
	if err := Set(root, None); err != nil {
		t.Fatal(err)
	}
	want := plain + "\n# slices\n[release]\nbranch = \"main\"\n"
	if got := fileOf(t, root); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := currentName(t, root); got != None {
		t.Fatalf("got %q", got)
	}
	// Nothing to remove is not an error.
	if err := Set(root, None); err != nil {
		t.Fatal(err)
	}
}

func TestAnUnknownArchitectureIsRefusedWithTheChoices(t *testing.T) {
	root := workspaceWith(t, plain)
	err := Set(root, "hexagonal")
	if err == nil || !strings.Contains(err.Error(), "vertical-slices, none") {
		t.Fatalf("%v", err)
	}
	if fileOf(t, root) != plain {
		t.Fatal("the file changed")
	}
}

func TestACommentedTableIsNotTheArchitectureInUse(t *testing.T) {
	root := workspaceWith(t, plain+"\n# [conventions.vertical-slices]\n")
	if got := currentName(t, root); got != None {
		t.Fatalf("got %q", got)
	}
}

func TestTheCatalogNamesEveryChoiceOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range Catalog() {
		if a.Name == "" || a.Meaning == "" || a.Writes() == "" || seen[a.Name] {
			t.Fatalf("bad entry %+v", a)
		}
		seen[a.Name] = true
	}
	if !seen[None] || !seen["vertical-slices"] {
		t.Fatalf("the first choices are missing: %v", Names())
	}
}
