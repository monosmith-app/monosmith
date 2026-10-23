package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateUsesTheGoToolchain(t *testing.T) {
	ws := t.TempDir()
	var calls []string
	GoRunner = func(dir string, argv []string) error {
		calls = append(calls, strings.Join(argv, " "))
		if argv[1] == "mod" {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+argv[3]+"\n"), 0o644)
		}
		return nil
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "my-sdk", Directory: "packages/my-sdk", Workspace: protocol.WorkspaceInfo{Name: "acme", ModulePrefix: "github.com/acme/repo"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"go work init", "go mod init github.com/acme/repo/packages/my-sdk", "go work use ./packages/my-sdk"}
	if strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("calls %v", calls)
	}
	goMod, _ := os.ReadFile(filepath.Join(ws, "packages", "my-sdk", "go.mod"))
	if !strings.HasPrefix(string(goMod), "// smith:kind package\n") {
		t.Fatalf("go.mod %q", goMod)
	}
	doc, _ := os.ReadFile(filepath.Join(ws, "packages", "my-sdk", "doc.go"))
	if !strings.Contains(string(doc), "package mysdk") {
		t.Fatalf("doc.go %q", doc)
	}
	if len(res.Files) < 5 {
		t.Fatalf("files %v", res.Files)
	}
}

func TestPackageName(t *testing.T) {
	for in, want := range map[string]string{"my-sdk": "mysdk", "UI": "ui", "3d": "pkg3d"} {
		if got := PackageName(in); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}

// go mod init writes the version of the toolchain that runs smith and go work
// use raises go.work to it, so a module added to a workspace declares the
// version the workspace has instead.
func TestGenerateKeepsTheWorkspacesGoVersion(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.26.0\n\nuse (\n\t./apps/api\n)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls []string
	GoRunner = func(dir string, argv []string) error {
		calls = append(calls, strings.Join(argv, " "))
		if argv[1] == "mod" && argv[2] == "init" {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+argv[3]+"\n\ngo 1.27.0\n"), 0o644)
		}
		return nil
	}
	_, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "actions", Directory: "libs/actions", Workspace: protocol.WorkspaceInfo{Name: "acme", ModulePrefix: "github.com/acme/repo"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"go mod init github.com/acme/repo/libs/actions", "go mod edit -go=1.26.0", "go work use ./libs/actions"}
	if strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("calls %v, want %v", calls, want)
	}
}

func TestWorkspaceGoVersionReadsTheGoLineOfGoWork(t *testing.T) {
	ws := t.TempDir()
	if got := WorkspaceGoVersion(ws); got != "" {
		t.Errorf("no go.work: %q", got)
	}
	if err := os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.26.0\r\n\r\nuse ./a\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := WorkspaceGoVersion(ws); got != "1.26.0" {
		t.Errorf("got %q", got)
	}
}
