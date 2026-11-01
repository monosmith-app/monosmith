package graphview

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

func TestHasEdges(t *testing.T) {
	g := &projectgraph.Graph{Projects: []*projectgraph.Project{{}, {}}}
	if hasEdges(g) {
		t.Fatal("no project has a dependency")
	}
	g.Projects[1].Deps = []string{"core"}
	if !hasEdges(g) {
		t.Fatal("a dependency is an edge")
	}
}
