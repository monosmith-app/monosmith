package cipipeline

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
)

// taskOf is the task of a task id ("core:test" is "test").
func taskOf(id string) string {
	_, task := projectOf(id)
	return task
}

// MarkedForVerify lists, by task, the projects whose task a plugin marked
// for verify, leaving out native projects, the verify tasks themselves and
// tasks whose program is not installed (named in missing, for a warning).
func MarkedForVerify(projects []*projectgraph.Project, installed func(string) bool) (byTask map[string][]string, missing []string) {
	byTask = map[string][]string{}
	for _, p := range projects {
		if slices.Contains(p.Tags, release.NativeTag) {
			continue
		}
		tasks := make([]string, 0, len(p.Tasks))
		for name := range p.Tasks {
			tasks = append(tasks, name)
		}
		sort.Strings(tasks)
		for _, name := range tasks {
			t := p.Tasks[name]
			if !t.Verify || slices.Contains(VerifyTasks, name) || len(t.Command) == 0 {
				continue
			}
			if program := t.Command[0]; !strings.HasPrefix(program, "{") && !installed(program) {
				missing = append(missing, fmt.Sprintf("%s:%s not verified: %s is not installed", p.Name, name, program))
				continue
			}
			byTask[name] = append(byTask[name], p.Name)
		}
	}
	return byTask, missing
}
