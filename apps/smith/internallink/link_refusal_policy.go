package internallink

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Refuse says why from cannot depend on lib, or returns nil when it can. The
// message names the project and the file or folder, and the fix. The rules
// are, in order: not itself, only a lib or a package can be depended upon,
// the same language, and no cycle.
func Refuse(g *projectgraph.Graph, from, lib *projectgraph.Project) error {
	switch {
	case from.Name == lib.Name:
		return fmt.Errorf("%s cannot depend on itself (%s). Name another library of the workspace", from.Name, from.Root)
	case lib.Kind == protocol.ProjectApp:
		return fmt.Errorf("%s is an app (%s), and an app cannot be depended upon. Move the shared code into a lib, or change its kind with smith kind", lib.Name, lib.Root)
	case from.Language != lib.Language:
		return fmt.Errorf("%s is %s and %s is %s: a project can link only a library of its own language. Use a registry package, or write the library in %s", from.Name, from.Language, lib.Name, lib.Language, from.Language)
	}
	if path := pathBetween(g, lib.Name, from.Name); path != nil {
		cycle := append([]string{from.Name}, path...)
		return fmt.Errorf("linking %s to %s makes a cycle: %s. Break it first, for example by moving what both need into a third library", from.Name, lib.Name, strings.Join(cycle, " → "))
	}
	return nil
}

// pathBetween returns the names from start along dependencies to goal, both
// included, or nil when start does not reach goal.
func pathBetween(g *projectgraph.Graph, start, goal string) []string {
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
		p := g.Get(cur)
		if p == nil {
			continue
		}
		for _, d := range p.Deps {
			if _, seen := prev[d]; !seen {
				prev[d] = cur
				queue = append(queue, d)
			}
		}
	}
	return nil
}
