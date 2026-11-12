package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// manifestHas fails the test unless the file at rel under ws holds want.
func manifestHas(t *testing.T, ws, rel, want string) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), want) {
		t.Fatalf("%s does not hold %q:\n%s", rel, want, src)
	}
}

func TestGenerateEndsTheModulePathWithThePackageName(t *testing.T) {
	ws := t.TempDir()
	real := GoRunner
	defer func() { GoRunner = real }()
	GoRunner = func(dir string, argv []string) error {
		if argv[1] == "mod" {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+argv[3]+"\n"), 0o644)
		}
		return nil
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "acme", ModulePrefix: "github.com/acme/repo"}, PackageName: "pargic"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/go.mod", "module github.com/acme/repo/libs/pargic\n")
	manifestHas(t, ws, "libs/core/doc.go", "package pargic")
}
