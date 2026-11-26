package taskrun

import "github.com/monosmith-app/monosmith/apps/smith/features/taskgraph"

// AttachedTask is the task of a run that gets smith's terminal (Spec.Terminal):
// an interactive program such as a TUI, a REPL or a dev server with key
// shortcuts then draws and reads keys. A task gets it only when it is the one
// task the person asked for (its dependencies run before it, prefixed as
// always), smith's input and output are a terminal, and nothing needs its
// output: no cache entry, no quiet run, no lint fix and no served app.
//
// Parameters: g is the run's graph; terminal says smith's standard input and
// output are a terminal; quiet is Options.Quiet.
// Returns: the task's ID, or "" when no task gets the terminal.
func AttachedTask(g *taskgraph.Graph, terminal, quiet bool) string {
	if !terminal || quiet || g == nil {
		return ""
	}
	var chosen *taskgraph.Task
	for _, id := range g.Order {
		t := g.Tasks[id]
		if !t.Requested {
			continue
		}
		if chosen != nil {
			return ""
		}
		chosen = t
	}
	if chosen == nil || chosen.Def.Cache || chosen.Fixing || chosen.Def.Serve != "" {
		return ""
	}
	return chosen.ID
}
