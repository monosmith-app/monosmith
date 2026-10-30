package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCheckPackageImports(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "sdk")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"acme-sdk\"\ndependencies = [\"PyYAML\"]\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "sdk.py"), []byte("import yaml\nimport numpy\n"), 0o644)

	p := protocol.Project{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage}
	f := CheckPackageImports(ws, p)
	if f.OK || f.Detail != "sdk.py imports numpy, which pyproject.toml does not declare" {
		t.Fatalf("finding %+v", f)
	}

	_ = os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"acme-sdk\"\ndependencies = [\"PyYAML\", \"numpy\"]\n"), 0o644)
	if f := CheckPackageImports(ws, p); !f.OK {
		t.Fatalf("declared: %+v", f)
	}

	_ = os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project\n"), 0o644)
	if f := CheckPackageImports(ws, p); f.OK || f.Detail == "" {
		t.Fatalf("an unreadable manifest is a failed check with the reason: %+v", f)
	}
}
