package scaffolding

import (
	"strings"
	"testing"
)

func TestWorkspaceEdition(t *testing.T) {
	for name, tc := range map[string]struct{ src, want string }{
		"none":                {"[workspace]\nmembers = []\n", ""},
		"shared":              {"[workspace]\nmembers = []\n\n[workspace.package]\nedition = \"2024\"\n", "2024"},
		"table without key":   {"[workspace.package]\nlicense = \"MIT\"\n\n[workspace.dependencies]\nedition = \"2018\"\n", ""},
		"a new workspace has": {NewWorkspace("libs/x"), DefaultEdition},
	} {
		if got := WorkspaceEdition(tc.src); got != tc.want {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
	}
}

func TestShareEditionMovesTheMajorityEditionToTheRoot(t *testing.T) {
	crates := map[string]string{
		"libs/a": "[package]\nname = \"a\"\nedition = \"2021\"\n\n[dependencies]\n",
		"libs/b": "[package]\nname = \"b\"\nedition = \"2021\" # keep\n",
		"libs/c": "[package]\nname = \"c\"\nedition = \"2018\"\n",
	}
	root, changed, edition := ShareEdition("[workspace]\nmembers = [\"libs/a\"]\n", crates)
	if edition != "2021" || !strings.HasSuffix(root, "\n[workspace.package]\nedition = \"2021\"\n") {
		t.Fatalf("root %q edition %q", root, edition)
	}
	if len(changed) != 2 || !strings.Contains(changed["libs/a"], "edition.workspace = true\n\n[dependencies]") || strings.Contains(changed["libs/b"], "2021") {
		t.Fatalf("changed %v", changed)
	}
	if _, ok := changed["libs/c"]; ok {
		t.Fatal("a crate with another edition keeps its own")
	}
}

func TestShareEditionKeepsTheRootsEditionAndFillsAnEmptyTable(t *testing.T) {
	root, changed, edition := ShareEdition("[workspace]\nmembers = []\n\n[workspace.package]\nlicense = \"MIT\"\n", map[string]string{"libs/a": "[package]\nname = \"a\"\nedition = \"2024\"\n", "libs/b": "[package]\nname = \"b\"\nedition = \"2024\"\n"})
	if edition != "2024" || !strings.Contains(root, "[workspace.package]\nedition = \"2024\"\nlicense") || len(changed) != 2 {
		t.Fatalf("root %q edition %q changed %v", root, edition, changed)
	}
	if _, changed, edition := ShareEdition("[workspace]\n", map[string]string{"libs/a": "[package]\nname = \"a\"\nedition = \"2021\"\n"}); edition != "" || changed != nil {
		t.Fatalf("one crate has nothing to share with: %q %v", edition, changed)
	}
	if _, changed, edition := ShareEdition("[workspace.package]\nedition = \"2024\"\n", map[string]string{"libs/a": "[package]\nedition.workspace = true\n"}); edition != "" || changed != nil {
		t.Fatalf("nothing to do: %q %v", edition, changed)
	}
}
