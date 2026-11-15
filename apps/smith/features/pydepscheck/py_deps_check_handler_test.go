package pydepscheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandFailsOnAnUndeclaredImport(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("smith.toml", "schema = 1\n[workspace]\nname = \"acme\"\n")
	write("packages/ocr/pyproject.toml", "[project]\nname = \"acme-ocr\"\ndependencies = [\"PyYAML>=6\"]\n")
	write("packages/ocr/ocr/__init__.py", "")
	write("packages/ocr/ocr/read.py", "import os\nimport yaml\nimport numpy\nfrom ocr import read\n")
	// smith run sets it for its tasks, this test among them.
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	t.Chdir(filepath.Join(ws, "packages", "ocr"))

	var out bytes.Buffer
	cmd := Command()
	// smith's root command silences these; Execute here runs this one alone.
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"packages/ocr"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "imports 1 module(s)") {
		t.Fatalf("err = %v", err)
	}
	if out.String() != "ocr/read.py imports numpy, which pyproject.toml does not declare\n" {
		t.Errorf("out = %q", out.String())
	}

	write("packages/ocr/pyproject.toml", "[project]\nname = \"acme-ocr\"\ndependencies = [\"PyYAML>=6\", \"numpy\"]\n")
	cmd = Command()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"packages/ocr"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("declared: %v", err)
	}
}
