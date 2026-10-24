package convertplan

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// planBuildOutputs makes each Node project's declared build outputs match
// where its build script writes. smith's default is the workspace's
// dist/<project dir>, but the converted scripts write where package.json's
// main and types point (the project's own dist): left alone, a cache hit
// would restore the declared folder and leave the real one empty. The
// project declares what its script writes as the build task's outputs.
func planBuildOutputs(plan *Plan, w *nxworkspace.Workspace) {
	for _, pp := range plan.Projects {
		if pp.Language != "node" {
			continue
		}
		script := pp.Scripts[buildoutputs.Task]
		if script == "" {
			for _, p := range w.Projects {
				if p.Root == pp.Root && p.Package != nil {
					script = p.Package.Scripts[buildoutputs.Task]
				}
			}
		}
		task := plan.Override(pp.Name).Tasks[buildoutputs.Task]
		if script == "" || len(task.Command) > 0 {
			continue
		}
		declared, fromConfig := task.Outputs, len(task.Outputs) > 0
		if !fromConfig {
			declared = []string{protocol.BuildDirFromProject(pp.Root)}
		}
		gap, found := buildoutputs.GapOf(pp.Name, pp.Root, script, declared, fromConfig)
		if !found {
			continue
		}
		task.Outputs = gap.Outputs
		plan.SetTask(pp.Name, buildoutputs.Task, task)
		plan.Add(Changed, pp.Name, buildoutputs.Task, "its build writes "+strings.Join(gap.Written, ", ")+" in the project, so those are the task's outputs ("+strings.Join(gap.Outputs, ", ")+"): a cache hit restores them")
	}
}
