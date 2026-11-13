package convertplan

import (
	"path"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// viteConfigs are the file names of a Vite config.
var viteConfigs = []string{"vite.config.ts", "vite.config.mts", "vite.config.js", "vite.config.mjs", "vite.config.cjs"}

// planBuildOutputs makes each Node project's build outputs match where its
// build script writes, with the detector smith doctor uses (buildoutputs).
//
// A build task whose outputs come from a plugin's default (the workspace's
// dist/<project dir>) takes what the script writes: the default is smith's,
// not the user's. A build task whose outputs the old tool declared (Nx
// outputs, Turborepo outputs) keeps them when the script writes inside them.
// When the script writes elsewhere, or a declared folder is never written,
// the plan takes the detected outputs and records an OutputChoice: convert
// asks before it applies, and --yes takes the detected ones.
func planBuildOutputs(plan *Plan, w *nxworkspace.Workspace, files []string, read func(string) string) {
	for _, pp := range plan.Projects {
		if pp.Language != "node" {
			continue
		}
		scripts := projectScripts(pp, w)
		task := plan.Override(pp.Name).Tasks[buildoutputs.Task]
		script, ok := buildScript(pp, task, scripts, w)
		if !ok {
			continue
		}
		if len(task.Outputs) == 0 {
			planDefaultOutputs(plan, pp, task, script)
			continue
		}
		expanded := buildoutputs.ExpandScript(script, scripts)
		written, complete := buildoutputs.ScriptWrites(expanded, viteNamesOutDir(pp.Root, files, read))
		if len(written) == 0 {
			continue
		}
		declared := cleanAll(task.Outputs)
		detected := buildoutputs.Suggested(written, declared)
		if !complete {
			// Something else in the script writes where its config says: keep
			// every declared folder and add only what is missing.
			detected = buildoutputs.Collapse(append(slices.Clone(declared), buildoutputs.Uncovered(written, declared)...))
		}
		if sameSet(detected, declared) {
			continue
		}
		task.Outputs = detected
		plan.SetTask(pp.Name, buildoutputs.Task, task)
		plan.OutputChoices = append(plan.OutputChoices, OutputChoice{Project: pp.Name, Task: buildoutputs.Task, Declared: declared, Detected: detected})
		plan.Decide(pp.Name, buildoutputs.Task+" outputs", outputChoiceMessage(plan.ToolName(), written, declared, detected))
	}
}

// planDefaultOutputs replaces the plugin's default outputs with what the
// build script writes. Only a task the plugin runs has the default; a command
// without outputs caches nothing and is left alone.
func planDefaultOutputs(plan *Plan, pp *ProjectPlan, task workspace.TaskOverride, script string) {
	if len(task.Command) > 0 {
		return
	}
	gap, found := buildoutputs.GapOf(pp.Name, pp.Root, script, []string{protocol.BuildDirFromProject(pp.Root)}, false)
	if !found {
		return
	}
	task.Outputs = gap.Outputs
	plan.SetTask(pp.Name, buildoutputs.Task, task)
	plan.Add(Changed, pp.Name, buildoutputs.Task, "its build writes "+strings.Join(gap.Written, ", ")+" in the project, so those are the task's outputs ("+strings.Join(gap.Outputs, ", ")+"): a cache hit restores them")
}

// outputChoiceMessage says what the build writes, what the old tool declared,
// and the two options.
func outputChoiceMessage(tool string, written, declared, detected []string) string {
	msg := "the build writes " + strings.Join(buildoutputs.Collapse(written), ", ") + ", but " + tool + " declared " + strings.Join(declared, ", ") + " as its outputs."
	if unwritten := unwrittenOutputs(written, declared); len(unwritten) > 0 {
		msg += " A cache hit restores only declared folders, and smith doctor fails a declared folder the build never writes (" + strings.Join(unwritten, ", ") + ")."
	} else {
		msg += " A cache hit restores only declared folders, so the rest would stay empty."
	}
	return msg + " Suggested: cache " + strings.Join(detected, ", ") + ". Or keep the declared outputs: convert asks before it applies, and --yes takes the suggestion"
}

// unwrittenOutputs lists the declared outputs that hold no written folder and
// lie in none.
func unwrittenOutputs(written, declared []string) []string {
	var out []string
	for _, d := range declared {
		used := false
		for _, w := range written {
			if buildoutputs.Covered(w, []string{d}) || buildoutputs.Covered(d, []string{w}) {
				used = true
			}
		}
		if !used {
			out = append(out, d)
		}
	}
	return out
}

// buildScript returns the package.json script the project's build task runs:
// build when the plugin runs it, or the script an `npm run <script>` command
// names, run in the project or in the workspace root with -w <package>.
func buildScript(pp *ProjectPlan, task workspace.TaskOverride, scripts map[string]string, w *nxworkspace.Workspace) (string, bool) {
	if len(task.Command) == 0 {
		return scripts[buildoutputs.Task], scripts[buildoutputs.Task] != ""
	}
	name, pkg, ok := buildoutputs.RunScript(task.Command)
	if !ok {
		return "", false
	}
	cwd := task.Cwd
	if cwd == "" {
		cwd = pp.Root
	}
	switch {
	case pkg == "" && path.Clean(cwd) == path.Clean(pp.Root):
	case pkg != "" && path.Clean(cwd) == "." && (pkg == packageName(pp, w) || pkg == pp.Root):
	default:
		return "", false
	}
	return scripts[name], scripts[name] != ""
}

// projectScripts are the project's package.json scripts as the conversion
// leaves them: the originals, with the scripts the plan writes over them.
func projectScripts(pp *ProjectPlan, w *nxworkspace.Workspace) map[string]string {
	scripts := map[string]string{}
	for _, p := range w.Projects {
		if p.Root == pp.Root && p.Package != nil {
			for k, v := range p.Package.Scripts {
				scripts[k] = v
			}
		}
	}
	for k, v := range pp.Scripts {
		scripts[k] = v
	}
	return scripts
}

// packageName is the name in the project's package.json.
func packageName(pp *ProjectPlan, w *nxworkspace.Workspace) string {
	for _, p := range w.Projects {
		if p.Root == pp.Root && p.Package != nil {
			return p.Package.Name
		}
	}
	return ""
}

// viteNamesOutDir reports whether the project's Vite config names an outDir,
// so a `vite build` writes there and not to Vite's default dist.
func viteNamesOutDir(root string, files []string, read func(string) string) bool {
	for _, name := range viteConfigs {
		rel := path.Join(root, name)
		if slices.Contains(files, rel) && strings.Contains(read(rel), "outDir") {
			return true
		}
	}
	return false
}

func cleanAll(dirs []string) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = path.Clean(d)
	}
	return out
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		if !slices.Contains(b, x) {
			return false
		}
	}
	return true
}
