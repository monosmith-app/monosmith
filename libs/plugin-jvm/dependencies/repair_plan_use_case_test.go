package dependencies

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRepairPurgesMavenDependenciesOnly(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"apps/api", "apps/android"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "api", "pom.xml"), []byte("<project/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "apps", "android", "build.gradle.kts"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	gradle := protocol.RepairParams{Projects: []protocol.Project{{Name: "android", Root: "apps/android"}}}
	if res, _ := Repair(root, gradle); len(res.Plans) != 0 {
		t.Fatalf("Gradle has no repair: %+v", res.Plans)
	}
	res, _ := Repair(root, protocol.RepairParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "android", Root: "apps/android"}}})
	if len(res.Plans) != 1 {
		t.Fatalf("one plan: %+v", res.Plans)
	}
	p := res.Plans[0]
	if p.Kind != protocol.RepairCache || !p.MachineWide || p.Install || len(p.Commands) != 1 || p.Commands[0].Dir != "apps/api" || p.Commands[0].Argv[2] != "dependency:purge-local-repository" {
		t.Fatalf("cache: %+v", p)
	}
}
