package discovery

import (
	"path"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestPackageLintDependsOnDepsCheckOnce guards against the deps-check task
// being added twice to a package's lint task.
func TestPackageLintDependsOnDepsCheckOnce(t *testing.T) {
	tasks := Tasks("libs/core", protocol.ProjectPackage, "core", ToolArgs{})
	n := 0
	for _, d := range tasks["lint"].DependsOn {
		if d == DepsCheckTask {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("lint depends on %s %d times, want 1: %v", DepsCheckTask, n, tasks["lint"].DependsOn)
	}
}

func TestTheToolsSkipTheSdistStagingFolder(t *testing.T) {
	has := func(cmd []string, flag, value string) bool {
		for i := 0; i+1 < len(cmd); i++ {
			if cmd[i] == flag && cmd[i+1] == value {
				return true
			}
		}
		return false
	}
	for root, pytestGlob := range map[string]string{"libs/core": "{projectRoot}/" + StagingGlob, ".": "{workspaceRoot}/" + StagingGlob} {
		tasks := Tasks(root, protocol.ProjectPackage, "core", ToolArgs{})
		if !has(tasks["lint"].Command, "--extend-exclude", StagingGlob) || !has(tasks["lint"].Fix.Command, "--extend-exclude", StagingGlob) {
			t.Errorf("%s: ruff skips the staging folder: %v", root, tasks["lint"].Command)
		}
		if !has(tasks["test"].Command, "--ignore-glob", pytestGlob) {
			t.Errorf("%s: pytest skips the staging folder of this project only: %v", root, tasks["test"].Command)
		}
	}
}

func TestStagingGlobMatchesAVersionFolderAndNoModule(t *testing.T) {
	for name, want := range map[string]bool{
		"monosmith-0.0.0": true, "my_pkg-1.2": true, "acme-2.0.0rc1": true,
		"core.py": false, "migrate-2024.py": false, "test_core.py": false, "src": false,
	} {
		if got, _ := path.Match(StagingGlob, name); got != want {
			t.Errorf("%q: got %v, want %v", name, got, want)
		}
	}
}
