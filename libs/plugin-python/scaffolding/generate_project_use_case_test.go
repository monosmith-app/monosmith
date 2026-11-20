package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"slices"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	var calls []string
	venv.Runner = func(_ string, argv []string) error {
		calls = append(calls, strings.Join(argv, " "))
		if len(argv) > 3 && argv[len(argv)-3] == "venv" {
			py := venv.Path(ws)
			_ = os.MkdirAll(filepath.Dir(py), 0o755)
			return os.WriteFile(py, nil, 0o755)
		}
		return nil
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "api", Directory: "apps/api", Workspace: protocol.WorkspaceInfo{Name: "Acme"}})
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Files, ",")
	for _, want := range []string{"apps/api/pyproject.toml", "apps/api/src/acme_api/__main__.py", "apps/api/src/acme_api/greeting/greet_use_case.py", "requirements-dev.txt"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %s in %v", want, res.Files)
		}
	}
	if _, err := os.Stat(filepath.Join(ws, "ruff.toml")); err == nil {
		t.Error("ruff's defaults are smith's: no ruff.toml is scaffolded")
	}
	pyproject, _ := os.ReadFile(filepath.Join(ws, "apps", "api", "pyproject.toml"))
	if !strings.Contains(string(pyproject), `name = "acme-api"`) || !strings.Contains(string(pyproject), `kind = "app"`) {
		t.Fatalf("pyproject %s", pyproject)
	}
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "-m venv --upgrade-deps .venv") || !strings.Contains(joined, "-r requirements-dev.txt") || !strings.Contains(joined, "-e ./apps/api") {
		t.Fatalf("calls %v", calls)
	}

	calls = nil
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: protocol.WorkspaceInfo{Name: "Acme"}}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(calls, "\n"), "-m venv") {
		t.Fatal("the venv must be created once")
	}
}

func TestGenerateManifestOnlyFindsFlatSources(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "libs", "sorts"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "sorts", Directory: "libs/sorts", ManifestOnly: true, Workspace: protocol.WorkspaceInfo{Name: "acme"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Files, []string{"libs/sorts/pyproject.toml"}) {
		t.Fatalf("want pyproject.toml only, got %v", res.Files)
	}
	raw, _ := os.ReadFile(filepath.Join(ws, "libs", "sorts", "pyproject.toml"))
	if !strings.Contains(string(raw), `where = ["."]`) {
		t.Errorf("a folder without src/ finds its packages at its top:\n%s", raw)
	}
	if _, err := os.Stat(filepath.Join(ws, ".venv")); err == nil {
		t.Error("ManifestOnly made a venv")
	}
}
