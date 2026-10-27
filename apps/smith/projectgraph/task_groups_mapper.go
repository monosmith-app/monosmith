package projectgraph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// InTaskGroup reports whether the project matches every selector of the group.
func (p *Project) InTaskGroup(g workspace.TaskGroup) bool {
	if len(g.Select) == 0 {
		return false
	}
	for _, selector := range g.Select {
		if !p.Matches(strings.TrimSpace(selector)) {
			return false
		}
	}
	return true
}

// TaskGroupSource is how a group is named where a value's origin is shown.
func TaskGroupSource(g workspace.TaskGroup) string {
	quoted := make([]string, len(g.Select))
	for i, s := range g.Select {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "smith.toml [[task_groups]] select = [" + strings.Join(quoted, ", ") + "]"
}

// applyTaskGroups layers the [[task_groups]] that name the project, in file
// order, so a later group wins field by field. A group only changes tasks the
// project already has: it describes how a set of projects runs what they
// share, and a task one of them lacks is not added.
func (p *Project) applyTaskGroups(cfg workspace.Config) {
	for _, g := range cfg.TaskGroups {
		if !p.InTaskGroup(g) {
			continue
		}
		names := make([]string, 0, len(g.Tasks))
		for name := range g.Tasks {
			if _, ok := p.Tasks[name]; ok {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			p.override(name, g.Tasks[name], TaskGroupSource(g))
		}
	}
}
