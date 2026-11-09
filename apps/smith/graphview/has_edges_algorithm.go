package graphview

import "github.com/monosmith-app/monosmith/apps/smith/projectgraph"

// hasEdges reports whether any project of g depends on another one.
func hasEdges(g *projectgraph.Graph) bool {
	for _, p := range g.Projects {
		if len(p.Deps) > 0 {
			return true
		}
	}
	return false
}
