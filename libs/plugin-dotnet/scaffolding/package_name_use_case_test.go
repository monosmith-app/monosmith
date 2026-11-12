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

func TestGenerateNamesTheProjectAfterThePackageName(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "acme"}, PackageName: "Pargic.Core"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/Pargic.Core.csproj", "")
}

// The default name starts with the workspace scope without its @, not the
// workspace name, as npm names do.
func TestGenerateBuildsTheProjectIDFromTheScope(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "opencraw", Scope: "@pargic"}}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/Pargic.Core.csproj", "")
}
