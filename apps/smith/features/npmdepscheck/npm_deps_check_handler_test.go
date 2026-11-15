package npmdepscheck

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
	write("package.json", `{"private":true,"workspaces":["packages/*"]}`)
	write("packages/ocr/package.json", `{"name":"@acme/ocr","dependencies":{"zod":"^4.0.0"}}`)
	write("packages/ocr/src/read.ts", "import { z } from 'zod'\nimport type { Page } from '@acme/core'\n")
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
	if out.String() != "src/read.ts imports @acme/core, which package.json does not declare\n" {
		t.Errorf("out = %q", out.String())
	}

	write("packages/ocr/package.json", `{"name":"@acme/ocr","dependencies":{"zod":"^4.0.0"},"peerDependencies":{"@acme/core":"^1.0.0"}}`)
	cmd = Command()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"packages/ocr"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("declared as a peer: %v", err)
	}
}
