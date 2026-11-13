package taskrun

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// overlapOrder finds the tasks of one run whose outputs overlap. A task
// clears its output folders before it runs and a cache hit replaces them, so
// two such tasks at once fail each other mid-write. Each task of a pair waits
// for the other when that one comes first in the run's order: they run one
// after the other. A folder under shared ([workspace] shared_outputs) is
// filled on purpose and never cleared, so it is not compared.
//
// Returns, per task, the tasks it waits for, and the warning for each pair.
func overlapOrder(g *taskgraph.Graph, shared []string) (map[string][]string, []string) {
	at := map[string]int{}
	var tasks []buildoutputs.TaskOutputs
	for i, id := range g.Order {
		t := g.Tasks[id]
		at[id] = i
		tasks = append(tasks, buildoutputs.TaskOutputs{ID: id, Name: t.Name, Root: t.Project.Root, Outputs: t.Def.Outputs})
	}
	after := map[string][]string{}
	var warnings []string
	for _, o := range buildoutputs.Overlaps(tasks, shared) {
		first, last := o.A, o.B
		if at[first] > at[last] {
			first, last = last, first
		}
		after[last] = append(after[last], first)
		warnings = append(warnings, fmt.Sprintf("%s %s and %s both write %s: smith runs them one after the other, and the folder keeps only the files of %s, the last one\n  → give each task its own folder (its command and its outputs), or run them in separate runs",
			marks.Warning, first, last, o.Where(), last))
	}
	return after, warnings
}
