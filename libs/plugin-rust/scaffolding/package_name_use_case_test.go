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

func TestGenerateNamesTheCrateAfterThePackageName(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "pargic-core", Directory: "libs/pargic-core", Workspace: protocol.WorkspaceInfo{Name: "opencraw"}, PackageName: "pargic"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/pargic-core/Cargo.toml", `name = "pargic"`)
}

// The default name starts with the workspace scope without its @, not the
// workspace name, as npm names do.
func TestGenerateBuildsTheCrateNameFromTheScope(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "opencraw", Namespace: "pargic"}}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/Cargo.toml", `name = "pargic-core"`)
}
