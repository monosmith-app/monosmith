package internallink

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Node is the part of a project the link rules read. The TUI builds nodes from
// its project list, the command from the project graph.
type Node struct {
	Name, Root, Kind, Language string
	// Deps are the names of the projects it depends on.
	Deps []string
}

// NodesOf turns the projects of a graph into nodes.
func NodesOf(g *projectgraph.Graph) []Node {
	out := make([]Node, 0, len(g.Projects))
	for _, p := range g.Projects {
		out = append(out, Node{Name: p.Name, Root: p.Root, Kind: string(p.Kind), Language: p.Language, Deps: p.Deps})
	}
	return out
}

// Refuse says why from cannot depend on lib, or returns nil when it can. The
// message names the project and the file or folder, and the fix. The rules
// are, in order: not itself, only a lib or a package can be depended upon,
// the same language, and no cycle.
func Refuse(g *projectgraph.Graph, from, lib *projectgraph.Project) error {
	return RefuseNodes(NodesOf(g), from.Name, lib.Name)
}

// RefuseNodes is Refuse over nodes, by project name.
func RefuseNodes(nodes []Node, from, lib string) error {
	index := map[string]Node{}
	for _, n := range nodes {
		index[n.Name] = n
	}
	f, l := index[from], index[lib]
	switch {
	case f.Name == l.Name:
		return fmt.Errorf("%s cannot depend on itself (%s). Name another library of the workspace", f.Name, f.Root)
	case l.Kind == string(protocol.ProjectApp):
		return fmt.Errorf("%s is an app (%s), and an app cannot be depended upon. Move the shared code into a lib, or change its kind with smith kind", l.Name, l.Root)
	case f.Language != l.Language:
		return fmt.Errorf("%s is %s and %s is %s: a project can link only a library of its own language. Use a registry package, or write the library in %s", f.Name, f.Language, l.Name, l.Language, f.Language)
	}
	if path := pathBetween(index, l.Name, f.Name); path != nil {
		cycle := append([]string{f.Name}, path...)
		return fmt.Errorf("linking %s to %s makes a cycle: %s. Break it first, for example by moving what both need into a third library", f.Name, l.Name, strings.Join(cycle, " → "))
	}
	return nil
}

// ValidLibraries lists the nodes that from may link: not itself, not an app,
// the same language, and no cycle, in the order given.
func ValidLibraries(nodes []Node, from string) []Node {
	var out []Node
	for _, n := range nodes {
		if RefuseNodes(nodes, from, n.Name) == nil {
			out = append(out, n)
		}
	}
	return out
}

// pathBetween returns the names from start along dependencies to goal, both
// included, or nil when start does not reach goal.
func pathBetween(index map[string]Node, start, goal string) []string {
	prev := map[string]string{start: ""}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == goal {
			var path []string
			for n := goal; n != ""; n = prev[n] {
				path = append([]string{n}, path...)
			}
			return path
		}
		for _, d := range index[cur].Deps {
			if _, seen := prev[d]; !seen {
				prev[d] = cur
				queue = append(queue, d)
			}
		}
	}
	return nil
}
