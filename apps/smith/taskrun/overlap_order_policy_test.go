package taskrun

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func graphOf(project *projectgraph.Project, outputs map[string][]string, order ...string) *taskgraph.Graph {
	g := &taskgraph.Graph{Tasks: map[string]*taskgraph.Task{}}
	for _, name := range order {
		id := project.Name + ":" + name
		g.Tasks[id] = &taskgraph.Task{ID: id, Name: name, Project: project, Def: protocol.Task{Outputs: outputs[name]}}
		g.Order = append(g.Order, id)
	}
	return g
}

func TestTwoTasksOfOneRunThatWriteOneFolderRunOneAfterTheOther(t *testing.T) {
	api := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api"}}
	outputs := map[string][]string{"build": {"dist"}, "build-dev": {"dist"}, "lint": nil}
	after, warnings := overlapOrder(graphOf(api, outputs, "build-dev", "lint", "build"), nil)
	if len(after) != 1 || strings.Join(after["api:build"], ",") != "api:build-dev" {
		t.Fatalf("api:build waits for api:build-dev, which comes first: %v", after)
	}
	want := "▲ api:build-dev and api:build both write apps/api/dist: smith runs them one after the other, and the folder keeps only the files of api:build, the last one\n  → give each task its own folder"
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], want) {
		t.Fatalf("warnings: %q", warnings)
	}

	// One of them alone, or both in a shared folder, is no warning.
	if after, warnings := overlapOrder(graphOf(api, outputs, "build"), nil); len(after) != 0 || len(warnings) != 0 {
		t.Errorf("one task: %v %v", after, warnings)
	}
	if after, warnings := overlapOrder(graphOf(api, outputs, "build", "build-dev"), []string{"apps/api/dist"}); len(after) != 0 || len(warnings) != 0 {
		t.Errorf("a shared folder: %v %v", after, warnings)
	}
}
