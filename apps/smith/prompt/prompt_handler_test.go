package prompt

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func run(t *testing.T, args ...string) string {
	t.Helper()
	cmd := Command("1.2.3")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// workspaceAt makes a folder with a smith.toml and points smith at it, so the
// test does not depend on the folder it runs in.
func workspaceAt(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SMITH_WORKSPACE_ROOT", root)
	return root
}

func TestOutsideAWorkspaceItPrintsNothingAndSucceeds(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", t.TempDir())
	if got := run(t); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := run(t, "--json"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestWithoutASavedRunItPrintsTheVersionOnly(t *testing.T) {
	workspaceAt(t)
	if got := run(t); got != "smith 1.2.3\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWithASavedRunItPrintsTheLineAndTheJSON(t *testing.T) {
	root := workspaceAt(t)
	err := Save(root, Summary{Schema: schemaVersion, Workspace: root, Failed: 1, Warnings: 4,
		Projects: []Project{{Name: "api", Root: "api"}, {Name: "web", Root: "web"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := run(t), "smith 1.2.3 · 2 projects · × 1 ▲ 4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	var v View
	if err := json.Unmarshal([]byte(run(t, "--json")), &v); err != nil {
		t.Fatal(err)
	}
	if v.Version != "1.2.3" || *v.Projects != 2 || v.Doctor.Failed != 1 || v.Doctor.Warnings != 4 {
		t.Fatalf("got %+v", v)
	}
}
