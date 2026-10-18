package cipipeline

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(name string, tags []string, tasks map[string]protocol.Task) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Tags: tags, Tasks: tasks}}
}

func TestMarkedForVerifyListsTheMarkedTasksThatCanRun(t *testing.T) {
	image := protocol.Task{Command: []string{"docker", "build", "."}, Verify: true}
	projects := []*projectgraph.Project{
		project("api", nil, map[string]protocol.Task{"image": image, "build": {Command: []string{"go", "build"}, Verify: true}}),
		project("web", nil, map[string]protocol.Task{"image": {Command: []string{"docker", "build", "."}}}),
		project("driver", []string{release.NativeTag}, map[string]protocol.Task{"image": image}),
		project("host", nil, map[string]protocol.Task{"image": image, "check": {Command: []string{"{smith}", "x"}, Verify: true}}),
	}
	byTask, missing := MarkedForVerify(projects, func(string) bool { return true })
	want := map[string][]string{"image": {"api", "host"}, "check": {"host"}}
	if !reflect.DeepEqual(byTask, want) || len(missing) != 0 {
		t.Errorf("byTask %v, missing %v", byTask, missing)
	}
	byTask, missing = MarkedForVerify(projects, func(p string) bool { return p != "docker" })
	if len(byTask["image"]) != 0 || len(missing) != 2 || !strings.Contains(missing[0], "api:image not verified: docker is not installed") {
		t.Errorf("without docker: byTask %v, missing %v", byTask, missing)
	}
}
