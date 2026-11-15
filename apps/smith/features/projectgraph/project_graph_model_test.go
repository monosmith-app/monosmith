package projectgraph

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func proj(name, root string, exports, imports []string) *Project {
	return &Project{Project: protocol.Project{Name: name, Root: root, Exports: exports, Imports: imports, Tasks: map[string]protocol.Task{}}, Plugin: "test"}
}

func sample(t *testing.T) *Graph {
	t.Helper()
	ps := []*Project{
		proj("web", "apps/web", []string{"@acme/web"}, []string{"@acme/ui/button", "react"}),
		proj("ui", "libs/ui", []string{"@acme/ui"}, []string{"@acme/core"}),
		proj("core", "libs/core", []string{"@acme/core"}, nil),
		proj("api", "apps/api", []string{"example.com/r/apps/api"}, []string{"example.com/r/libs/gocore/greeting"}),
		proj("gocore", "libs/gocore", []string{"example.com/r/libs/gocore"}, nil),
	}
	ResolveEdges(ps)
	g, err := New(ps)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestEdgesOrderAndWalks(t *testing.T) {
	g := sample(t)
	if !reflect.DeepEqual(g.Get("web").Deps, []string{"ui"}) || !reflect.DeepEqual(g.Get("api").Deps, []string{"gocore"}) {
		t.Fatalf("deps web=%v api=%v", g.Get("web").Deps, g.Get("api").Deps)
	}
	order, err := g.Order()
	if err != nil {
		t.Fatal(err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if pos["core"] > pos["ui"] || pos["ui"] > pos["web"] || pos["gocore"] > pos["api"] {
		t.Fatalf("order %v", order)
	}
	if !reflect.DeepEqual(g.TransitiveDeps("web"), []string{"core", "ui"}) || !reflect.DeepEqual(g.TransitiveDependents("core"), []string{"ui", "web"}) {
		t.Fatal("walks")
	}
	if g.Owner("libs/ui/src/x.ts").Name != "ui" || g.Owner("README.md") != nil || g.Owner("libs/uix/a") != nil {
		t.Fatal("owner")
	}
}

func TestCycleAndDuplicates(t *testing.T) {
	a := proj("a", "libs/a", []string{"a"}, []string{"b"})
	b := proj("b", "libs/b", []string{"b"}, []string{"a"})
	ResolveEdges([]*Project{a, b})
	g, _ := New([]*Project{a, b})
	if _, err := g.Order(); err == nil || !strings.Contains(err.Error(), "a → b → a") {
		t.Fatalf("cycle error %v", err)
	}
	if _, err := New([]*Project{proj("x", "apps/x", nil, nil), proj("x", "libs/x", nil, nil)}); err == nil {
		t.Fatal("duplicate names must fail")
	}
}

func TestApplyOverridesRecordsSources(t *testing.T) {
	p := proj("web", "apps/web", nil, nil)
	p.Tasks["test"] = protocol.Task{Command: []string{"jest"}, Cache: true}
	off := false
	cfg := workspace.Config{
		Tasks:    map[string]workspace.TaskOverride{"test": {Cache: &off}},
		Projects: map[string]workspace.ProjectOverride{"web": {Tasks: map[string]workspace.TaskOverride{"e2e": {Command: []string{"playwright", "test"}}, "test": {Inputs: []string{"src/**"}}}}},
	}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if p.Tasks["test"].Cache || p.Tasks["test"].Inputs[0] != "src/**" || p.Tasks["e2e"].Command[0] != "playwright" {
		t.Fatalf("tasks %+v", p.Tasks)
	}
	want := map[string]string{"test.cache": "smith.toml [tasks.test]", "test.inputs": "smith.toml [projects.web.tasks.test]", "test.command": "plugin test"}
	for k, v := range want {
		if p.Sources[k] != v {
			t.Errorf("%s: %q want %q", k, p.Sources[k], v)
		}
	}
	bad := workspace.Config{Projects: map[string]workspace.ProjectOverride{"web": {Tasks: map[string]workspace.TaskOverride{"nope": {}}}}}
	if err := ApplyOverrides(p, bad); err == nil {
		t.Fatal("an override of a missing task without a command must fail")
	}
}

func TestDeployersRelease(t *testing.T) {
	deploy := protocol.InitializeResult{Roles: []protocol.Role{protocol.RoleDeploy}}
	image := protocol.Contribution{Tasks: map[string]protocol.Task{"image": {}}}
	app := &Project{Project: protocol.Project{Kind: protocol.ProjectApp}}
	if app.Releases() {
		t.Fatal("an app releases only when it opts in or ships")
	}
	app.addDeployer("lint-only", protocol.InitializeResult{}, image)
	app.addDeployer("docker", deploy, protocol.Contribution{})
	if app.Releases() {
		t.Fatal("a plugin without the deploy role, or a deploy plugin that added nothing, ships nothing")
	}
	app.addDeployer("docker", deploy, image)
	app.addDeployer("docker", deploy, image)
	if !app.Releases() || len(app.Deployers) != 1 {
		t.Fatalf("an app a deploy plugin ships releases: %v", app.Deployers)
	}
	lib := &Project{Project: protocol.Project{Kind: protocol.ProjectLib}}
	lib.addDeployer("docker", deploy, image)
	if lib.Releases() {
		t.Fatal("a lib never releases")
	}
}
