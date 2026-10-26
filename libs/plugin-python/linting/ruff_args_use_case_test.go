package linting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, ws, rel, body string) {
	t.Helper()
	full := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWithoutAConfigSmithPassesItsDefault(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/api/pyproject.toml", "[project]\nname = \"api\"\n")
	args, own, err := Args(ws, "apps/api")
	if err != nil || own {
		t.Fatalf("%v %v", own, err)
	}
	if got := strings.Join(args, " "); got != "--config {workspaceRoot}/.smith/defaults/ruff.toml" {
		t.Fatalf("args: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "ruff.toml"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestTheRepositorysRuffConfigWins(t *testing.T) {
	cases := map[string]string{
		"ruff.toml":               "line-length = 80\n",
		".ruff.toml":              "line-length = 80\n",
		"pyproject.toml":          "[tool.ruff]\nline-length = 80\n",
		"apps/api/pyproject.toml": "[project]\nname = \"api\"\n\n[tool.ruff.lint]\nselect = [\"E\"]\n",
		"apps/api/.ruff.toml":     "line-length = 80\n",
	}
	for where, body := range cases {
		ws := t.TempDir()
		write(t, ws, "apps/api/pyproject.toml", "[project]\nname = \"api\"\n")
		write(t, ws, where, body)
		args, own, err := Args(ws, "apps/api")
		if err != nil || !own || len(args) != 0 {
			t.Errorf("%s: %v %v %v", where, args, own, err)
		}
		if _, err := os.Stat(filepath.Join(ws, ".smith")); err == nil {
			t.Errorf("%s: no default is written when the repository has its own", where)
		}
	}
}

func TestAPyprojectWithoutRuffDoesNotCount(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[tool.pytest.ini_options]\ntestpaths = [\"tests\"]\n")
	if _, own, _ := Args(ws, "apps/api"); own {
		t.Fatal("only a [tool.ruff] table configures ruff")
	}
}
