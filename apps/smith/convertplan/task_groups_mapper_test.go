package convertplan

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

func sharedTestLib(t *testing.T, name, env string) *nxworkspace.Project {
	return project(name, "libs/"+name, "lib", []string{"package.json"}, map[string]nxworkspace.Target{
		"test": target(t, `{"executor": "nx:run-commands", "dependsOn": ["^build", "logger:build"], "options": {"cwd": "libs/`+name+`", "command": "jest", "env": {"MODE": "`+env+`"}}}`),
	})
}

// A workspace of libraries that Nx configured the same way gets the shared
// settings once, as a task group; the one that differs keeps its own.
func TestPlanWritesSettingsSharedByLibrariesAsATaskGroup(t *testing.T) {
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{
		sharedTestLib(t, "a", "ci"), sharedTestLib(t, "b", "ci"), sharedTestLib(t, "c", "ci"), sharedTestLib(t, "odd", "local"),
	}}
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})

	if len(plan.Config.TaskGroups) != 1 {
		t.Fatalf("groups = %+v; projects = %+v", plan.Config.TaskGroups, plan.Config.Projects)
	}
	g := plan.Config.TaskGroups[0]
	if strings.Join(g.Select, ",") != "lang:node,kind:lib" {
		t.Errorf("select = %v", g.Select)
	}
	if got := g.Tasks["test"].Env["MODE"]; got != "ci" {
		t.Errorf("group env = %v", g.Tasks["test"].Env)
	}
	for _, name := range []string{"a", "b", "c"} {
		if o, ok := plan.Config.Projects[name]; ok && len(o.Tasks) > 0 {
			t.Errorf("%s still repeats %v", name, o.Tasks)
		}
	}
	odd := plan.Config.Projects["odd"].Tasks["test"]
	if odd.Env["MODE"] != "local" || strings.Join(odd.DependsOn, ",") != strings.Join(g.Tasks["test"].DependsOn, ",") {
		t.Errorf("odd = %+v, group = %+v; it keeps its env and gets the group's other fields", odd, g.Tasks["test"])
	}
	if mapped := findings(plan, Mapped); !strings.Contains(mapped, "test: 3 projects share the same test settings, written once as a task group for lang:node and kind:lib (a, b, c)") {
		t.Errorf("mapped:\n%s", mapped)
	}
}

func TestPlanKeepsOverridesThatOnlyOneProjectHas(t *testing.T) {
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{sharedTestLib(t, "a", "ci"), project("b", "libs/b", "lib", []string{"package.json"}, nil)}}
	plan := MakePlan(w, Input{DirName: "acme", Read: func(string) string { return "" }})
	if len(plan.Config.TaskGroups) != 0 {
		t.Errorf("groups = %+v", plan.Config.TaskGroups)
	}
	if len(plan.Config.Projects["a"].Tasks) == 0 {
		t.Error("a lost its override")
	}
}
