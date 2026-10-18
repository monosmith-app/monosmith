package affected

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWatchPathsAreTheAppsAndItsTransitiveDependencies(t *testing.T) {
	mk := func(name, root string, kind protocol.ProjectKind, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root, Kind: kind}, Deps: deps}
	}
	g, err := projectgraph.New([]*projectgraph.Project{
		mk("web", "apps/web", protocol.ProjectApp, "ui"),
		mk("api", "apps/api", protocol.ProjectApp, "core"),
		mk("ui", "libs/ui", protocol.ProjectLib, "core"),
		mk("core", "libs/core", protocol.ProjectLib),
		mk("unrelated", "libs/unrelated", protocol.ProjectLib),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := WatchPaths(g)
	want := map[string][]string{
		"web": {"apps/web", "libs/core", "libs/ui"},
		"api": {"apps/api", "libs/core"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
}
