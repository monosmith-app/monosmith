package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
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

func TestGenerateNamesTheDistributionAfterThePackageName(t *testing.T) {
	ws := t.TempDir()
	real := venv.Runner
	defer func() { venv.Runner = real }()
	venv.Runner = func(_ string, argv []string) error {
		if len(argv) > 3 && argv[len(argv)-3] == "venv" {
			py := venv.Path(ws)
			_ = os.MkdirAll(filepath.Dir(py), 0o755)
			return os.WriteFile(py, nil, 0o755)
		}
		return nil
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "Acme"}, PackageName: "pargic-core"}); err != nil {
		t.Fatal(err)
	}
	manifestHas(t, ws, "libs/core/pyproject.toml", `name = "pargic-core"`)
	manifestHas(t, ws, "libs/core/src/pargic_core/__init__.py", "")
}
