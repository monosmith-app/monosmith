package taskgraph

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// ServeTarget resolves a task's serve setting ("web:start", or "start" in the
// task's own project) to the project and task it names.
func ServeTarget(pg *projectgraph.Graph, project, serve string) (*projectgraph.Project, string, error) {
	name, task := project, serve
	if before, after, found := strings.Cut(serve, ":"); found {
		name, task = before, after
	}
	p := pg.Get(name)
	if p == nil {
		return nil, "", fmt.Errorf("serve %q: there is no project %q", serve, name)
	}
	if _, ok := p.Tasks[task]; !ok {
		return nil, "", fmt.Errorf("serve %q: %s has no %q task", serve, name, task)
	}
	return p, task, nil
}
