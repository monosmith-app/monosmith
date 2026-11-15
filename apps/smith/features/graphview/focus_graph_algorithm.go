package graphview

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
)

// Around returns the part of the graph that touches the named projects: them,
// everything they depend on and everything that depends on them, transitively.
func Around(g *projectgraph.Graph, names []string) *projectgraph.Graph {
	keep := map[string]bool{}
	for _, n := range names {
		keep[n] = true
		for _, d := range g.TransitiveDeps(n) {
			keep[d] = true
		}
		for _, d := range g.TransitiveDependents(n) {
			keep[d] = true
		}
	}
	return Subgraph(g, keep)
}

// Subgraph keeps the projects in keep, and only the edges between them.
func Subgraph(g *projectgraph.Graph, keep map[string]bool) *projectgraph.Graph {
	var projects []*projectgraph.Project
	for _, p := range g.Projects {
		if !keep[p.Name] {
			continue
		}
		copied := *p
		copied.Deps = nil
		for _, d := range p.Deps {
			if keep[d] {
				copied.Deps = append(copied.Deps, d)
			}
		}
		sort.Strings(copied.Deps)
		projects = append(projects, &copied)
	}
	sub, _ := projectgraph.New(projects)
	return sub
}
