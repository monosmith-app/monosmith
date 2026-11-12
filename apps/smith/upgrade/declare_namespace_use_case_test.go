package upgrade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func namespaceSession(t *testing.T, toml string, ws workspace.WorkspaceSection) (*session.Session, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	return &session.Session{Root: dir, Config: workspace.Config{Workspace: ws}}, dir
}

func TestUpgradeKeepsTheNameTheScopeGave(t *testing.T) {
	s, dir := namespaceSession(t, "[workspace]\nname = \"monosmith\"\nscope = \"@acme\"\n\n[ci]\n", workspace.WorkspaceSection{Name: "monosmith", Scope: "@acme"})
	var r Result
	if err := declareNamespace(s, true, &r); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, "smith.toml")); strings.Contains(string(raw), "namespace") {
		t.Fatal("a check changes nothing")
	}
	if len(r.Migrations) != 1 || !strings.Contains(r.Migrations[0], `namespace = "acme"`) || !strings.Contains(r.Migrations[0], "monosmith-core") {
		t.Fatalf("%v", r.Migrations)
	}
	r = Result{}
	if err := declareNamespace(s, false, &r); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "smith.toml"))
	if want := "[workspace]\nname = \"monosmith\"\nscope = \"@acme\"\nnamespace = \"acme\"\n\n[ci]\n"; string(raw) != want {
		t.Fatalf("%q", raw)
	}
}

func TestUpgradeLeavesAWorkspaceWhoseNamesAreUnchanged(t *testing.T) {
	for name, ws := range map[string]workspace.WorkspaceSection{
		"no scope":         {Name: "acme"},
		"scope is name":    {Name: "Acme", Scope: "@acme"},
		"namespace is set": {Name: "monosmith", Scope: "@acme", Namespace: "pargic"},
	} {
		s, _ := namespaceSession(t, "[workspace]\n", ws)
		var r Result
		if err := declareNamespace(s, false, &r); err != nil || len(r.Migrations) != 0 || len(r.Files) != 0 {
			t.Errorf("%s: %v %+v", name, err, r)
		}
	}
}
