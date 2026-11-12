package discovery

import (
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
