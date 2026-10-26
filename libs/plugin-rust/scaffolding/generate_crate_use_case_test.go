package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateMakesAWorkspaceThenAddsToIt(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "cli", Directory: "apps/cli", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info, Stack: "rust"})
	if err != nil {
		t.Fatal(err)
	}
	root, _ := os.ReadFile(filepath.Join(ws, "Cargo.toml"))
	if string(root) != "[workspace]\nresolver = \"2\"\nmembers = [\"apps/cli\", \"packages/core\"]\n" {
		t.Fatalf("root Cargo.toml:\n%s", root)
	}
	manifest, _ := os.ReadFile(filepath.Join(ws, "packages", "core", "Cargo.toml"))
	crate, err := discovery.ParseCargo(manifest)
	if err != nil || crate.Name != "acme-core" || crate.Kind != "package" || crate.Stack != "rust" {
		t.Fatalf("%v %+v\n%s", err, crate, manifest)
	}
	if strings.Contains(string(manifest), "publish = false") {
		t.Fatal("a package is published")
	}
	for _, f := range []string{"apps/cli/src/main.rs", "apps/cli/src/greeting/greet_use_case.rs", "packages/core/src/lib.rs"} {
		if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(f))); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	if config, _ := os.ReadFile(filepath.Join(ws, ".cargo", "config.toml")); !strings.Contains(string(config), `target-dir = "dist/cargo"`) {
		t.Fatalf("Cargo builds into dist/:\n%s", config)
	}
	if !strings.Contains(strings.Join(res.Files, ","), "Cargo.toml") {
		t.Fatalf("files %v", res.Files)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "cli", Directory: "apps/cli", Workspace: info}); err == nil {
		t.Fatal("an existing directory is refused")
	}
	out, err := Remove(ws, protocol.RemoveParams{Project: protocol.Project{Root: "apps/cli"}})
	root, _ = os.ReadFile(filepath.Join(ws, "Cargo.toml"))
	if err != nil || len(out.Files) != 1 || strings.Contains(string(root), "apps/cli") {
		t.Fatalf("remove: %v %+v\n%s", err, out, root)
	}
}

func TestMembers(t *testing.T) {
	multi := "[workspace]\nmembers = [\n    \"a\",\n]\n\n[workspace.dependencies]\nserde = \"1\"\n"
	got, err := AddMember(multi, "b")
	if err != nil || got != "[workspace]\nmembers = [\n    \"a\",\n    \"b\",\n]\n\n[workspace.dependencies]\nserde = \"1\"\n" {
		t.Fatalf("%v\n%s", err, got)
	}
	if again, _ := AddMember(got, "b"); again != got {
		t.Fatal("adding twice changes nothing")
	}
	if got, _ := AddMember("[workspace]\nresolver = \"2\"\n", "a"); got != "[workspace]\nmembers = [\"a\"]\nresolver = \"2\"\n" {
		t.Fatalf("no members yet:\n%s", got)
	}
	if _, err := AddMember("[package]\nname = \"x\"\n", "a"); err == nil {
		t.Fatal("no workspace table")
	}
}
func TestCrateNameIsValidForCargo(t *testing.T) {
	for _, c := range []struct{ workspace, name, want string }{
		{"acme", "core", "acme-core"},
		{"jato.auto.packages", "rust-core", "jato-auto-packages-rust-core"},
		{"My Shop!", "api", "my-shop-api"},
		{"acme", "acme-core", "acme-core"},
		{"jato.auto", "jato-auto-core", "jato-auto-core"},
		{"", "core", "core"},
	} {
		if got := CrateName(c.workspace, c.name); got != c.want {
			t.Errorf("CrateName(%q, %q) = %q, want %q", c.workspace, c.name, got, c.want)
		}
	}
}

func TestGenerateNoteHasNoPlaceholder(t *testing.T) {
	res, err := Generate(t.TempDir(), protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: protocol.WorkspaceInfo{Name: "a.b"}})
	if err != nil {
		t.Fatal(err)
	}
	note := strings.Join(res.Notes, "\n")
	if strings.Contains(note, "<") || !strings.Contains(note, `a-b-core = { path = "../../packages/core" }`) {
		t.Fatalf("note: %s", note)
	}
}
