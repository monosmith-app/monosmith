package godepscheck

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
	write("go.work", "go 1.24\n\nuse ./packages/ocr\n")
	write("packages/ocr/go.mod", "module example.com/acme/ocr\n\ngo 1.24\n\nrequire github.com/spf13/cobra v1.10.2\n")
	write("packages/ocr/read.go", "package ocr\n\nimport (\n\t\"fmt\"\n\n\t\"github.com/spf13/cobra\"\n\t\"golang.org/x/text/language\"\n)\n")
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
	if err == nil || !strings.Contains(err.Error(), "imports 1 package(s)") {
		t.Fatalf("err = %v", err)
	}
	if out.String() != "read.go imports golang.org/x/text/language, which go.mod does not require\n" {
		t.Errorf("out = %q", out.String())
	}

	write("packages/ocr/go.mod", "module example.com/acme/ocr\n\ngo 1.24\n\nrequire (\n\tgithub.com/spf13/cobra v1.10.2\n\tgolang.org/x/text v0.30.0\n)\n")
	cmd = Command()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"packages/ocr"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("required: %v", err)
	}
}
