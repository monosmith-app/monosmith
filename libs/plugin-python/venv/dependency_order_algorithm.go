package venv

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// DependencyOrder returns the projects with each one after the projects it
// imports (a library before the app that lists it in its dependencies), and
// otherwise in the order given. pip then finds a library of the workspace
// already installed in the venv when it installs the project that needs it,
// and never looks the name up on PyPI. A cycle does not loop: its projects come out in some fixed order.
func DependencyOrder(projects []protocol.Project) []protocol.Project {
	exporter := map[string]int{}
	for i, p := range projects {
		for _, e := range p.Exports {
			exporter[e] = i
		}
	}
	const (
		unseen = iota
		visiting
		done
	)
	state := make([]int, len(projects))
	out := make([]protocol.Project, 0, len(projects))
	var visit func(i int)
	visit = func(i int) {
		if state[i] != unseen {
			return
		}
		state[i] = visiting
		for _, imp := range projects[i].Imports {
			if j, ok := exporter[imp]; ok && j != i {
				visit(j)
			}
		}
		state[i] = done
		out = append(out, projects[i])
	}
	for i := range projects {
		visit(i)
	}
	return out
}
