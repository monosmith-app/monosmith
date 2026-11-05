package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func goProject(command ...string) []protocol.Project {
	return []protocol.Project{{Name: "cli", Root: "tools/cli", Tasks: map[string]protocol.Task{"build": {Command: command}}}}
}

func TestGoBuildTaskThatSmithMadeIsSilentAndAHandWrittenOneIsReported(t *testing.T) {
	if f := CheckBuildDirs(goProject("go", "build", "-trimpath", "-o", "{workspaceRoot}/dist/tools/cli/cli", ".")); len(f) != 0 {
		t.Fatalf("smith's own task passes -o: %+v", f)
	}
	f := CheckBuildDirs(goProject("go", "build", "-trimpath", "."))
	if len(f) != 1 || f[0].OK || !strings.Contains(f[0].Detail, "passes no -o") || !strings.Contains(f[0].Remedy, "-o {workspaceRoot}/dist/tools/cli/<binary>") {
		t.Fatalf("%+v", f)
	}
	if f := CheckBuildDirs(goProject("go", "build", "./...")); len(f) != 0 {
		t.Fatalf("go build ./... writes no binary: %+v", f)
	}
	if f := CheckBuildDirs(goProject("go", "build", "-o", "bin/cli", ".")); len(f) != 1 || !strings.Contains(f[0].Detail, "writes into tools/cli/bin/cli") {
		t.Fatalf("%+v", f)
	}
}
