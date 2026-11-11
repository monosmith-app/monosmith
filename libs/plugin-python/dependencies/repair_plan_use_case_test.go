package dependencies

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairRecreatesTheVenvAndSaysVersionsCanChange(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".venv"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Plans) != 1 {
		t.Fatalf("no interpreter in .venv: only reinstall, got %+v", res.Plans)
	}
	p := res.Plans[0]
	if p.Kind != protocol.RepairReinstall || len(p.Delete) != 1 || p.Delete[0] != ".venv" || !p.Install || !p.ChangesVersions {
		t.Fatalf("reinstall deletes .venv, installs, and versions can change: %+v", p)
	}
}

func TestRepairClearsPipsCacheWithTheVenvInterpreter(t *testing.T) {
	root := t.TempDir()
	python := venv.Path(root)
	if err := os.MkdirAll(filepath.Dir(python), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(python, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	res, _ := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}})
	if len(res.Plans) != 2 {
		t.Fatalf("want reinstall and cache, got %+v", res.Plans)
	}
	c := res.Plans[1]
	if c.Kind != protocol.RepairCache || !c.MachineWide || !c.Install || c.Commands[0].Argv[0] != python || c.Commands[0].Argv[4] != "purge" {
		t.Fatalf("cache: %+v", c)
	}
}

func TestRepairHasNoPlanWithoutProjects(t *testing.T) {
	if res, _ := Repair(t.TempDir(), protocol.RepairParams{}); len(res.Plans) != 0 {
		t.Fatalf("got %+v", res.Plans)
	}
}
