package testrunning

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
	if got := strings.Join(args, " "); got != "-c {workspaceRoot}/.smith/defaults/pytest.ini --rootdir=." {
		t.Fatalf("args: %s", got)
	}
	written, err := os.ReadFile(filepath.Join(ws, ".smith", "defaults", "pytest.ini"))
	if err != nil || string(written) != DefaultConfig() {
		t.Fatalf("the default is materialised: %q %v", written, err)
	}
}

func TestTheRepositorysPytestConfigWins(t *testing.T) {
	cases := map[string]string{
		"pytest.ini":              "[pytest]\naddopts = -x\n",
		".pytest.ini":             "[pytest]\naddopts = -x\n",
		"pyproject.toml":          "[tool.pytest.ini_options]\naddopts = \"-x\"\n",
		"tox.ini":                 "[pytest]\naddopts = -x\n",
		"setup.cfg":               "[tool:pytest]\naddopts = -x\n",
		"apps/api/pyproject.toml": "[project]\nname = \"api\"\n\n[tool.pytest]\naddopts = [\"-x\"]\n",
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

func TestAPyprojectWithoutPytestDoesNotCount(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[tool.ruff]\nline-length = 100\n")
	if _, own, _ := Args(ws, "apps/api"); own {
		t.Fatal("only a [tool.pytest] table configures pytest")
	}
}
