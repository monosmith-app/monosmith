package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	if res := CI(protocol.CIParams{}); len(res.Toolchains) != 0 {
		t.Error("no projects, no toolchain")
	}
	if res := CI(protocol.CIParams{Projects: []protocol.Project{{}}}); len(res.Toolchains) != 1 || res.Toolchains[0].Install == "" {
		t.Errorf("%+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "x"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "x", "x-scm-1.rockspec"), []byte("package = \"x\"\nversion = \"1.0.0-1\"\n"), 0o644)
	projects := protocol.DoctorParams{Projects: []protocol.Project{{Name: "x", Root: "libs/x"}}}

	probe = func(string, []string) (string, error) { return "", errors.New("missing") }
	res := Doctor(ws, projects)
	if len(res.Findings) != 3 || res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("without the tools, both are reported: %+v", res)
	}

	probe = func(string, []string) (string, error) { return "", nil }
	res = Doctor(ws, projects)
	name := res.Findings[2]
	if !res.Findings[0].OK || name.OK || !name.Warning || name.Fix == nil || len(name.Fix.Files) != 2 ||
		!name.Fix.Files[0].Delete || name.Fix.Files[0].Path != "libs/x/x-scm-1.rockspec" || name.Fix.Files[1].Path != "libs/x/x-1.0.0-1.rockspec" {
		t.Fatalf("%+v", res)
	}

	_ = os.Rename(filepath.Join(ws, "libs", "x", "x-scm-1.rockspec"), filepath.Join(ws, "libs", "x", "x-1.0.0-1.rockspec"))
	if res = Doctor(ws, projects); !res.Findings[2].OK {
		t.Errorf("a matching name passes: %+v", res.Findings[2])
	}
	if res = Doctor(ws, protocol.DoctorParams{}); len(res.Findings) != 0 {
		t.Errorf("no projects, no findings: %+v", res)
	}
}
