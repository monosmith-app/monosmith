package boundaries

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(name string, kind protocol.ProjectKind, tags []string, imports ...string) *projectgraph.Project {
	return &projectgraph.Project{
		Project: protocol.Project{Name: name, Root: "x/" + name, Kind: kind, Tags: tags, Exports: []string{name}, Imports: imports, Tasks: map[string]protocol.Task{}},
		Plugin:  "test",
	}
}

func graph(t *testing.T, ps ...*projectgraph.Project) *projectgraph.Graph {
	t.Helper()
	projectgraph.ResolveEdges(ps)
	g, err := projectgraph.New(ps)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func workspaceOf(t *testing.T) *projectgraph.Graph {
	return graph(t,
		project("web", protocol.ProjectApp, []string{"ui"}, "ui-kit", "admin"),
		project("admin", protocol.ProjectApp, []string{"ui"}),
		project("ui-kit", protocol.ProjectLib, []string{"ui"}, "server-core"),
		project("server-core", protocol.ProjectLib, []string{"server"}),
	)
}

func TestNoRulesOnlyWarnsAboutADependencyOnAnApp(t *testing.T) {
	vs := Check(workspaceOf(t), workspace.BoundariesSection{})
	if len(vs) != 1 || vs[0].Edge() != "web -> admin" || !vs[0].Warning {
		t.Fatalf("violations %+v, want one warning for web -> admin", vs)
	}
	if len(Errors(vs)) != 0 {
		t.Fatal("the unset built-in rule must not fail a check")
	}
}

func TestKindsTrueMakesItAnErrorAndFalseSwitchesItOff(t *testing.T) {
	on, off := true, false
	vs := Check(workspaceOf(t), workspace.BoundariesSection{Kinds: &on})
	if len(Errors(vs)) != 1 || !strings.Contains(vs[0].Rule, "app") || !strings.Contains(vs[0].Hint, "remove web's import of admin") {
		t.Fatalf("violations %+v", vs)
	}
	if vs := Check(workspaceOf(t), workspace.BoundariesSection{Kinds: &off}); len(vs) != 0 {
		t.Fatalf("switched off, got %+v", vs)
	}
}

func TestRuleByTagDeniesAndNamesTheReason(t *testing.T) {
	off := false
	cfg := workspace.BoundariesSection{Kinds: &off, Rules: []workspace.BoundaryRule{
		{From: "tag:ui", Deny: []string{"tag:server"}, Reason: "the browser must not pull in server code"},
	}}
	vs := Check(workspaceOf(t), cfg)
	if len(vs) != 1 || vs[0].Edge() != "ui-kit -> server-core" || vs[0].Warning || !strings.Contains(vs[0].Rule, "the browser must not pull in server code") {
		t.Fatalf("violations %+v", vs)
	}
}

func TestRuleByKindAllowsOnlyTheListed(t *testing.T) {
	off := false
	cfg := workspace.BoundariesSection{Kinds: &off, Rules: []workspace.BoundaryRule{{From: "kind:app", Allow: []string{"kind:lib"}}}}
	vs := Check(workspaceOf(t), cfg)
	if len(vs) != 1 || vs[0].Edge() != "web -> admin" || !strings.Contains(vs[0].Rule, `may only depend on "kind:lib"`) {
		t.Fatalf("violations %+v", vs)
	}
}

func TestAnExceptionClearsOneEdge(t *testing.T) {
	on := true
	cfg := workspace.BoundariesSection{Kinds: &on, Exceptions: []workspace.BoundaryException{{From: "web", To: "admin", Reason: "shared login page"}}}
	if vs := Check(workspaceOf(t), cfg); len(vs) != 0 {
		t.Fatalf("violations %+v", vs)
	}
}

func TestRemovingTheImportClearsTheViolation(t *testing.T) {
	on := true
	g := graph(t,
		project("web", protocol.ProjectApp, nil, "ui-kit"),
		project("admin", protocol.ProjectApp, nil),
		project("ui-kit", protocol.ProjectLib, nil),
	)
	if vs := Check(g, workspace.BoundariesSection{Kinds: &on}); len(vs) != 0 {
		t.Fatalf("violations %+v", vs)
	}
}
