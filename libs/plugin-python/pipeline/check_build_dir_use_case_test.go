package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPythonBuildTaskThatSmithMadeIsSilentAndAHandWrittenOneIsReported(t *testing.T) {
	made := discovery.Tasks("libs/kit", protocol.ProjectPackage, "kit", discovery.ToolArgs{})
	if f := CheckBuildDirs([]protocol.Project{{Name: "kit", Root: "libs/kit", Tasks: made}}); len(f) != 0 {
		t.Fatalf("smith's own task passes --outdir: %+v", f)
	}
	byHand := protocol.Project{Name: "kit", Root: "libs/kit", Tasks: map[string]protocol.Task{"build": {Command: []string{"python", "-m", "build", "."}}}}
	f := CheckBuildDirs([]protocol.Project{byHand})
	if len(f) != 1 || f[0].OK || !strings.Contains(f[0].Detail, "passes no --outdir") || !strings.Contains(f[0].Remedy, "--outdir {workspaceRoot}/dist/libs/kit") {
		t.Fatalf("%+v", f)
	}
}
