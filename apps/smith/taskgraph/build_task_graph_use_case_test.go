package taskgraph

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func graph(t *testing.T) *projectgraph.Graph {
	t.Helper()
	mk := func(name string, deps []string, tasks map[string]protocol.Task) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: "x/" + name, Tasks: tasks}, Deps: deps}
	}
	build := protocol.Task{Command: []string{"b"}, DependsOn: []string{"^build"}}
	test := protocol.Task{Command: []string{"t"}, DependsOn: []string{"^build"}}
	g, err := projectgraph.New([]*projectgraph.Project{
		mk("web", []string{"glue"}, map[string]protocol.Task{"build": build, "test": test}),
		// glue has no build: web's ^build must reach through it to core.
		mk("glue", []string{"core"}, map[string]protocol.Task{"test": test}),
		mk("core", nil, map[string]protocol.Task{"build": build, "gen": {Command: []string{"g"}}, "test": {Command: []string{"t"}, DependsOn: []string{"gen"}}}),
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestBuildBridgesAndOrders(t *testing.T) {
	g, err := Build(graph(t), []string{"web", "glue", "core"}, []string{"test"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g.Tasks["web:test"].Deps, []string{"core:build"}) {
		t.Fatalf("web:test deps %v", g.Tasks["web:test"].Deps)
	}
	if !reflect.DeepEqual(g.Tasks["core:test"].Deps, []string{"core:gen"}) {
		t.Fatalf("core:test deps %v", g.Tasks["core:test"].Deps)
	}
	if g.Tasks["core:build"].Requested || !g.Tasks["web:test"].Requested {
		t.Fatal("requested flags")
	}
	pos := map[string]int{}
	for i, id := range g.Order {
		pos[id] = i
	}
	if pos["core:build"] > pos["web:test"] || pos["core:gen"] > pos["core:test"] {
		t.Fatalf("order %v", g.Order)
	}
}

func TestBuildErrors(t *testing.T) {
	pg := graph(t)
	if _, err := Build(pg, []string{"nope"}, []string{"test"}); err == nil || !strings.Contains(err.Error(), "no project named") {
		t.Fatalf("got %v", err)
	}
	if _, err := Build(pg, []string{"glue"}, []string{"build"}); err == nil {
		t.Fatal("no task anywhere must fail")
	}
	core := pg.Get("core")
	core.Tasks["loop"] = protocol.Task{DependsOn: []string{"loop2"}}
	core.Tasks["loop2"] = protocol.Task{DependsOn: []string{"loop"}}
	if _, err := Build(pg, []string{"core"}, []string{"loop"}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("got %v", err)
	}
	core.Tasks["bad"] = protocol.Task{DependsOn: []string{"web:missing"}}
	if _, err := Build(pg, []string{"core"}, []string{"bad"}); err == nil {
		t.Fatal("dependency on a missing task must fail")
	}
}

func TestWhatTheServedTaskNeedsBuiltFirstIsADependencyOfTheServingTask(t *testing.T) {
	api := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"build"}},
		"start": {Command: []string{"start"}, DependsOn: []string{"build"}},
	}}}
	e2e := &projectgraph.Project{Project: protocol.Project{Name: "api-e2e", Root: "apps/api-e2e", Tasks: map[string]protocol.Task{
		"e2e": {Command: []string{"e2e"}, Serve: "api:start"},
	}}}
	pg, err := projectgraph.New([]*projectgraph.Project{api, e2e})
	if err != nil {
		t.Fatal(err)
	}
	g, err := Build(pg, []string{"api-e2e"}, []string{"e2e"})
	if err != nil {
		t.Fatal(err)
	}
	if got := g.Tasks["api-e2e:e2e"].Deps; !reflect.DeepEqual(got, []string{"api:build"}) {
		t.Errorf("deps %v", got)
	}
	if _, started := g.Tasks["api:start"]; started {
		t.Error("the served task is started by the runner, not run as a dependency")
	}
}
