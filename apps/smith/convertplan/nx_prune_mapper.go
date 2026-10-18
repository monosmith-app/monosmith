package convertplan

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// pruneTask is the node plugin's task that writes an app's deploy output.
const pruneTask = "prune"

// nxPruneExecutors write a Node app's deploy output, as the node plugin's
// prune task does: a production package.json and lockfile, and the
// workspace packages it uses in workspace_modules/.
var nxPruneExecutors = map[string]bool{
	"@nx/js:prune-lockfile":         true,
	"@nx/js:copy-workspace-modules": true,
	"@nx/node:prune":                true,
}

// IsNxPrune reports whether a target is one of Nx's prune steps: an executor
// above, or the prune target that only runs them (nx:noop).
func IsNxPrune(name string, t nxworkspace.Target) bool {
	return nxPruneExecutors[t.Executor] || (name == pruneTask && t.Executor == "nx:noop")
}

// mapNxPrune maps one of a Node app's Nx prune steps to the node plugin's
// prune task. It writes into the directory the app's build wrote to, so a
// Dockerfile written for Nx's output keeps working: when that is not
// smith's dist/<app dir>, the task is given it (--out).
func mapNxPrune(plan *Plan, p *nxworkspace.Project, pp *ProjectPlan, name string) {
	out := nxBuildOutput(p)
	reason := "the node plugin's prune task writes the same deploy output (smith run prune)"
	if out != "" && out != "dist/"+p.Root {
		if _, set := plan.Override(pp.Name).Tasks[pruneTask]; !set {
			plan.SetTask(pp.Name, pruneTask, workspace.TaskOverride{Command: []string{"{smith}", "npm-prune", p.Root, "--out", out}})
		}
		reason += ", into " + out + " where the build writes"
	}
	plan.Add(Mapped, pp.Name, name, reason)
}

// nxBuildOutput is the workspace-relative directory an Nx build target
// writes to (its outputPath option, or its first output), or "".
func nxBuildOutput(p *nxworkspace.Project) string {
	build, ok := p.Targets["build"]
	if !ok {
		return ""
	}
	var out string
	if !build.Option("outputPath", &out) || out == "" {
		if len(build.Outputs) == 0 {
			return ""
		}
		out = build.Outputs[0]
	}
	out = strings.TrimPrefix(strings.ReplaceAll(out, "{projectRoot}", p.Root), "{workspaceRoot}/")
	if strings.Contains(out, "{") {
		return ""
	}
	return strings.TrimSuffix(out, "/")
}
