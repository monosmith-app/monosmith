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

func TestGenerateNamesTheRockAfterThePackageName(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "strings", Directory: "libs/strings", Workspace: protocol.WorkspaceInfo{Name: "acme"}, PackageName: "pargic-strings"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/strings/pargic-strings-0.0.0-1.rockspec", `package = "pargic-strings"`)
}
