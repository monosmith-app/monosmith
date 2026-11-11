package taskoutput

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func crate(build protocol.Task) protocol.Project {
	return protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{"build": build}}
}

func TestCheckShippedOutputPassesWhenTheTaskCachesOnlyTheShippedFolder(t *testing.T) {
	for _, build := range []protocol.Task{
		{Outputs: []string{"../../dist/apps/api"}, Ship: []string{"../../dist/cargo/debug/api"}},
		{}, // a library crate ships nothing and caches nothing
	} {
		if f, ok := CheckShippedOutput(crate(build), "build", "dist/cargo"); ok {
			t.Errorf("%+v: %+v", build, f)
		}
	}
	if _, ok := CheckShippedOutput(protocol.Project{Name: "api"}, "build", "dist/cargo"); ok {
		t.Error("no build task")
	}
}

func TestCheckShippedOutputRefusesASharedWorkFolderAsAnOutput(t *testing.T) {
	for _, out := range []string{"../../dist/cargo", "../../dist/cargo/debug", "../../dist"} {
		f, ok := CheckShippedOutput(crate(protocol.Task{Outputs: []string{out}}), "build", "dist/cargo")
		if !ok || f.OK || f.Warning || f.Project != "api" || f.Check != "api caches only dist/apps/api" ||
			!strings.Contains(f.Detail, "empties each output folder first") || !strings.Contains(f.Detail, "builds into dist/cargo/") ||
			f.Remedy != "list only ../../dist/apps/api in the outputs of [projects.api.tasks.build] in smith.toml" {
			t.Errorf("%s: %+v", out, f)
		}
	}
}

func TestCheckShippedOutputAcceptsAWorkFolderTheToolDoesNotShare(t *testing.T) {
	jar := protocol.Task{Outputs: []string{"../../dist/apps/api", "../../dist/maven/api"}, Ship: []string{"../../dist/maven/api/*.jar"}}
	if f, ok := CheckShippedOutput(crate(jar), "build"); ok {
		t.Errorf("%+v", f)
	}
}

func TestCheckShippedOutputReportsAShippingTaskThatDoesNotCacheTheShippedFolder(t *testing.T) {
	build := protocol.Task{Ship: []string{"../../dist/cargo/debug/api"}}
	f, ok := CheckShippedOutput(crate(build), "build", "dist/cargo")
	if !ok || !f.Warning || f.Problem != "does not cache dist/apps/api" || strings.Contains(f.Detail, "cache hit") {
		t.Fatalf("an uncached task is a warning: %+v", f)
	}
	build.Cache = true
	if f, ok = CheckShippedOutput(crate(build), "build", "dist/cargo"); !ok || f.Warning || !strings.Contains(f.Detail, "a cache hit restores nothing") {
		t.Fatalf("a cached task is a failure: %+v", f)
	}
}
