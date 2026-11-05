package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDartBuildTaskIsReportedOnlyWhenItWritesOutsideDist(t *testing.T) {
	task := func(c ...string) protocol.Project {
		return protocol.Project{Name: "cli", Root: "apps/cli", Tasks: map[string]protocol.Task{"build": {Command: c}}}
	}
	for _, ok := range [][]string{
		{"dart", "compile", "exe", "bin/main.dart", "-o", "{workspaceRoot}/dist/apps/cli/cli"},
		{"flutter", "build", "web", "-o", "{workspaceRoot}/dist/apps/cli"},
		{"dart", "run", "build_runner", "build"},
	} {
		if f := CheckBuildDirs([]protocol.Project{task(ok...)}); len(f) != 0 {
			t.Errorf("%v: %+v", ok, f)
		}
	}
	f := CheckBuildDirs([]protocol.Project{task("flutter", "build", "web")})
	if len(f) != 1 || !strings.Contains(f[0].Detail, "passes no -o") || !strings.Contains(f[0].Remedy, "-o {workspaceRoot}/dist/apps/cli") {
		t.Fatalf("%+v", f)
	}
}
