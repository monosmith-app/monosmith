package typechecking

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
	args, own, err := Args(ws, "apps/api")
	if err != nil || own {
		t.Fatalf("%v %v", own, err)
	}
	if got := strings.Join(args, " "); got != "--config-file {workspaceRoot}/.smith/defaults/mypy.ini" {
		t.Fatalf("args: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "mypy.ini"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestTheRepositorysMypyConfigWins(t *testing.T) {
	cases := map[string]string{
		"mypy.ini":                "[mypy]\nstrict = True\n",
		".mypy.ini":               "[mypy]\nstrict = True\n",
		"pyproject.toml":          "[tool.mypy]\nstrict = true\n",
		"setup.cfg":               "[mypy]\nstrict = True\n",
		"apps/api/pyproject.toml": "[project]\nname = \"api\"\n\n[tool.mypy.overrides]\n",
	}
	for where, body := range cases {
		ws := t.TempDir()
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

func TestAPyprojectWithoutMypyDoesNotCount(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[tool.ruff]\nline-length = 100\n")
	if _, own, _ := Args(ws, "apps/api"); own {
		t.Fatal("only a [tool.mypy] table configures mypy")
	}
}
