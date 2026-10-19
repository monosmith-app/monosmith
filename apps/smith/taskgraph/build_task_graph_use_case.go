package taskgraph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Build creates the graph for running tasks on projects. A project that lacks
// a requested task is skipped silently (not every project has a typecheck);
// at least one task must exist overall.
func Build(pg *projectgraph.Graph, projects, tasks []string) (*Graph, error) {
	g := &Graph{Tasks: map[string]*Task{}}
	for _, name := range projects {
		p := pg.Get(name)
		if p == nil {
			return nil, fmt.Errorf("no project named %q (projects: %s)", name, strings.Join(pg.Names(), ", "))
		}
		for _, t := range tasks {
			if _, ok := p.Tasks[t]; !ok {
				continue
			}
			if err := add(g, pg, name, t); err != nil {
				return nil, err
			}
			g.Tasks[ID(name, t)].Requested = true
		}
	}
	if len(g.Tasks) == 0 {
		return nil, fmt.Errorf("none of the selected projects has a %s task", strings.Join(tasks, "/"))
	}
	order, err := topo(g)
	if err != nil {
		return nil, err
	}
	g.Order = order
	return g, nil
}

func add(g *Graph, pg *projectgraph.Graph, project, task string) error {
	id := ID(project, task)
	if _, ok := g.Tasks[id]; ok {
		return nil
	}
	p := pg.Get(project)
	def := p.Tasks[task]
	node := &Task{ID: id, Project: p, Name: task, Def: def}
	g.Tasks[id] = node
	deps := map[string]bool{}
	// The app a task serves is started by the runner, not run as a
	// dependency; what it needs built first is a dependency of this task.
	dependsOn := def.DependsOn
	depsOf := map[string]*projectgraph.Project{}
	if def.Serve != "" {
		sp, st, err := ServeTarget(pg, project, def.Serve)
		if err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		for _, d := range sp.Tasks[st].DependsOn {
			dependsOn = append(dependsOn[:len(dependsOn):len(dependsOn)], d)
			depsOf[d] = sp
		}
	}
	for _, dep := range dependsOn {
		p, project := p, project
		if owner := depsOf[dep]; owner != nil {
			p, project = owner, owner.Name
		}
		if strings.HasPrefix(dep, "^") {
			name := dep[1:]
			for _, target := range upstreamWith(pg, p.Deps, name) {
				if err := add(g, pg, target, name); err != nil {
					return err
				}
				deps[ID(target, name)] = true
			}
			continue
		}
		depProject, depTask := project, dep
		if before, after, found := strings.Cut(dep, ":"); found {
			depProject, depTask = before, after
		}
		dp := pg.Get(depProject)
		if dp == nil {
			return fmt.Errorf("%s depends on %s, but there is no project %q", id, dep, depProject)
		}
		if _, ok := dp.Tasks[depTask]; !ok {
			return fmt.Errorf("%s depends on %s, but %s has no %q task", id, dep, depProject, depTask)
		}
		if err := add(g, pg, depProject, depTask); err != nil {
			return err
		}
		deps[ID(depProject, depTask)] = true
	}
	for d := range deps {
		node.Deps = append(node.Deps, d)
	}
	sort.Strings(node.Deps)
	return nil
}

// upstreamWith returns, for each direct dependency, that project if it has the
// task, otherwise the nearest projects below it that do.
func upstreamWith(pg *projectgraph.Graph, direct []string, task string) []string {
	seen := map[string]bool{}
	var out []string
	var visit func(string)
	visit = func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		p := pg.Get(name)
		if p == nil {
			return
		}
		if _, ok := p.Tasks[task]; ok {
			out = append(out, name)
			return
		}
		for _, d := range p.Deps {
			visit(d)
		}
	}
	for _, d := range direct {
		visit(d)
	}
	sort.Strings(out)
	return out
}

func topo(g *Graph) ([]string, error) {
	ids := make([]string, 0, len(g.Tasks))
	for id := range g.Tasks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	state := map[string]int{}
	var order, stack []string
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 2:
			return nil
		case 1:
			i := 0
			for stack[i] != id {
				i++
			}
			return fmt.Errorf("tasks depend on each other in a cycle: %s", strings.Join(append(stack[i:], id), " → "))
		}
		state[id] = 1
		stack = append(stack, id)
		for _, d := range g.Tasks[id].Deps {
			if err := visit(d); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[id] = 2
		order = append(order, id)
		return nil
	}
	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return order, nil
}
