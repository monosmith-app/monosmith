package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorChecksWhatAPackageImports(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "packages", "sdk")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sdk\n\ngo 1.24\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "x.go"), []byte("package sdk\n\nimport _ \"github.com/spf13/cobra\"\n"), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.24\nuse ./packages/sdk\n"), 0o644)

	p := protocol.Project{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage}
	f := CheckPackageImports(ws, p)
	if f.OK || f.Detail != "x.go imports github.com/spf13/cobra, which go.mod does not require" || f.Fix != nil {
		t.Fatalf("finding %+v", f)
	}

	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sdk\n\ngo 1.24\n\nrequire github.com/spf13/cobra v1.10.2\n"), 0o644)
	if f := CheckPackageImports(ws, p); !f.OK {
		t.Fatalf("required: %+v", f)
	}

	// Doctor lists the check for a package and none for an app, and the go.work fix does not touch it.
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{p, {Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}}})
	found := 0
	for _, f := range res.Findings {
		if strings.HasSuffix(f.Check, "requires the modules it imports") {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("findings %+v", res.Findings)
	}
}
