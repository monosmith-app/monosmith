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

func TestGenerateNamesThePackageAfterThePackageName(t *testing.T) {
	ws := t.TempDir()
	real := Runner
	defer func() { Runner = real }()
	Runner = func(string, []string) error { return nil }
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "acme"}, PackageName: "pargic_core"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/pubspec.yaml", "name: pargic_core\n")
}

// The default name starts with the workspace scope without its @, not the
// workspace name, as npm names do.
func TestGenerateBuildsThePackageNameFromTheScope(t *testing.T) {
	ws := t.TempDir()
	real := Runner
	defer func() { Runner = real }()
	Runner = func(string, []string) error { return nil }
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "opencraw", Scope: "@pargic"}}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/pubspec.yaml", "name: pargic_core\n")
}
