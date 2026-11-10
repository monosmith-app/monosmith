package projectgraph

import (
	"reflect"
	"testing"
)

// A plugin reports a dependency its manifest declares among the project's
// imports (npm dependencies, go.mod requirements, Python requirements, Cargo
// path dependencies, project references), so a library linked by
// smith install is an edge before any source file imports it.
func TestADependencyOnlyTheManifestDeclaresIsAnEdge(t *testing.T) {
	ps := []*Project{
		proj("shop-web", "apps/shop-web", []string{"shop-web"}, []string{"@acme/core"}),
		proj("core", "libs/core", []string{"@acme/core"}, nil),
		proj("core-py", "libs/core-py", []string{"core_py"}, nil),
		proj("api", "apps/api", []string{"api"}, []string{"core_py", "requests"}),
	}
	ResolveEdges(ps)
	got := map[string][]string{}
	for _, p := range ps {
		got[p.Name] = p.Deps
	}
	want := map[string][]string{"shop-web": {"core"}, "core": {}, "core-py": {}, "api": {"core-py"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
