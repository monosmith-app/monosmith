package projectgraph

import (
	"sort"
	"strings"
)

// ResolveEdges sets each project's Deps: a project depends on another when it
// imports something the other exports, exactly or as a sub-path
// ("@acme/ui/button" imports "@acme/ui"; "example.com/r/libs/core/greeting"
// imports "example.com/r/libs/core").
func ResolveEdges(projects []*Project) {
	exporter := map[string]string{}
	for _, p := range projects {
		for _, e := range p.Exports {
			exporter[e] = p.Name
		}
	}
	for _, p := range projects {
		deps := map[string]bool{}
		for _, imp := range p.Imports {
			for candidate := imp; candidate != ""; candidate = parent(candidate) {
				if owner, ok := exporter[candidate]; ok {
					if owner != p.Name {
						deps[owner] = true
					}
					break
				}
			}
		}
		p.Deps = make([]string, 0, len(deps))
		for d := range deps {
			p.Deps = append(p.Deps, d)
		}
		sort.Strings(p.Deps)
	}
}

func parent(identifier string) string {
	i := strings.LastIndex(identifier, "/")
	if i <= 0 {
		return ""
	}
	return identifier[:i]
}
