package turboplan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// pluginDependsOn is what the Node plugin makes each well-known script depend
// on (libs/plugin-node/discovery); an override that says the same is left out
// of smith.toml.
var pluginDependsOn = map[string][]string{
	"build": {"^build"}, "typecheck": {"^build"}, "test": {"^build"}, "lint": {"^build"}, "start": {"build"},
}

// pluginCached are the scripts the Node plugin caches.
var pluginCached = map[string]bool{"build": true, "typecheck": true, "test": true, "lint": true}

// knownTaskKeys are the task fields this conversion reads or reports by name.
var knownTaskKeys = map[string]bool{
	"dependsOn": true, "outputs": true, "inputs": true, "env": true, "passThroughEnv": true,
	"cache": true, "persistent": true, "interactive": true, "outputLogs": true, "outputMode": true,
	"with": true, "extends": true, "description": true, "$schema": true,
}

// taskContext is what mapping one task of one package needs.
type taskContext struct {
	plan *convertplan.Plan
	w    *turboworkspace.Workspace
	// names maps package names to smith project names.
	names map[string]string
	// global are globalDependencies, as inputs.
	global []string
}

// effectiveTask is the definition Turborepo runs for a package's task: the
// package-specific root entry ("web#build") replaces the general one, and the
// package's own turbo.json is laid over either.
func effectiveTask(w *turboworkspace.Workspace, p *turboworkspace.Package, task string) (turboworkspace.TaskConfig, bool) {
	tasks := w.Turbo.TaskSet()
	def, ok := tasks[p.Name+"#"+task]
	if !ok {
		def, ok = tasks[task]
	}
	if p.Turbo != nil {
		if child, has := p.Turbo.TaskSet()[task]; has {
			if !ok {
				return child, true
			}
			return turboworkspace.MergeTask(def, child), true
		}
	}
	return def, ok
}

// mapPackageTasks carries the turbo.json settings of every script a package
// has into smith.toml, field by field, and reports what it cannot.
func mapPackageTasks(c taskContext, p *turboworkspace.Package, pp *convertplan.ProjectPlan) {
	scripts := make([]string, 0, len(p.JSON.Scripts))
	for s := range p.JSON.Scripts {
		scripts = append(scripts, s)
	}
	sort.Strings(scripts)
	for _, script := range scripts {
		def, ok := effectiveTask(c.w, p, script)
		if !ok {
			continue
		}
		o := workspace.TaskOverride{}
		var notes []string

		if deps, set := dependsOn(c, p, pp, script, def); set {
			o.DependsOn = deps
			notes = append(notes, "depends_on "+list(deps))
		}
		persistent := def.Persistent != nil && *def.Persistent
		off := def.Cache != nil && !*def.Cache
		turboCaches := !off && !persistent
		switch {
		case !turboCaches:
			no := false
			o.Cache = &no
			if off {
				notes = append(notes, "cache = false")
			}
		case !pluginCached[script]:
			// Turborepo caches every task unless it says otherwise; the
			// Node plugin caches only the well-known ones.
			yes := true
			o.Cache = &yes
			notes = append(notes, "cache = true")
		}
		if def.Outputs != nil {
			out := OutputDirs(def.Outputs)
			if out.Dirs != nil {
				o.Outputs = out.Dirs
				notes = append(notes, "outputs "+list(out.Dirs))
			}
			for _, ex := range out.Excluded {
				c.plan.Add(convertplan.Changed, pp.Name, script, fmt.Sprintf("the output exclusion %s cannot be stated (smith caches output directories whole), so the directory is cached with what the glob left out", ex))
			}
			for _, loose := range out.Loose {
				c.plan.Add(convertplan.Manual, pp.Name, script, fmt.Sprintf("the output %s is not a directory inside the package; smith caches directories, so list the directory it sits in under outputs in smith.toml", loose))
			}
			if script == "build" && len(out.Dirs) > 0 {
				c.plan.Add(convertplan.Changed, pp.Name, script, "the build writes to "+strings.Join(out.Dirs, ", ")+" in the package; smith's Node convention is the workspace's dist/"+pp.Root+", which smith doctor checks")
			}
		}
		if turboCaches && (def.Inputs != nil || len(c.global) > 0) {
			inputs := []string{"**/*"}
			if def.Inputs != nil {
				inputs = InputGlobs(def.Inputs)
			}
			inputs = append(inputs, c.global...)
			// Turborepo's hash covers the lockfile; so does the Node plugin's.
			inputs = append(inputs, "{workspaceRoot}/package-lock.json")
			o.Inputs = dedupe(inputs)
			notes = append(notes, "inputs "+list(o.Inputs))
		}
		if o.DependsOn != nil || o.Outputs != nil || o.Inputs != nil || o.Cache != nil {
			c.plan.SetTask(pp.Name, script, o)
			c.plan.Add(convertplan.Mapped, pp.Name, script, "turbo.json task kept in smith.toml: "+strings.Join(notes, ", "))
		}
	}
}

// dependsOn translates a task's dependsOn. set is false when the plugin's own
// dependencies already say the same.
func dependsOn(c taskContext, p *turboworkspace.Package, pp *convertplan.ProjectPlan, script string, def turboworkspace.TaskConfig) (deps []string, set bool) {
	deps = []string{}
	for _, d := range def.DependsOn {
		switch {
		case strings.HasPrefix(d, "$"):
			c.plan.Add(convertplan.Unsupported, pp.Name, script, "dependsOn "+d+" is a Turborepo 1 environment variable key; see env")
		case strings.HasPrefix(d, "^"):
			deps = append(deps, d)
		case strings.HasPrefix(d, "//#"):
			c.plan.Add(convertplan.Manual, pp.Name, script, "it runs after the root task "+d+"; smith runs projects, not the workspace root, so run that script before smith run")
		case strings.Contains(d, "#"):
			pkg, task, _ := strings.Cut(d, "#")
			name, known := c.names[pkg]
			switch {
			case !known:
				c.plan.Add(convertplan.Manual, pp.Name, script, "it runs after "+d+", and "+pkg+" is not a package of the workspace; add the dependency by hand")
			case pkg == p.Name:
				if hasScript(p, task) {
					deps = append(deps, task)
				}
			default:
				deps = append(deps, name+":"+task)
			}
		default:
			// Turborepo skips a dependency the package has no script for.
			if hasScript(p, d) {
				deps = append(deps, d)
			}
		}
	}
	want, known := pluginDependsOn[script]
	if known && script == "start" && !hasScript(p, "build") {
		want = nil
	}
	if equal(deps, want) {
		return nil, false
	}
	return deps, true
}

func hasScript(p *turboworkspace.Package, name string) bool {
	_, ok := p.JSON.Scripts[name]
	return ok
}

// reportTaskDefinition says what a task definition holds that smith has no
// place for. subject names where the definition is.
func reportTaskDefinition(plan *convertplan.Plan, subject string, def turboworkspace.TaskConfig) {
	if len(def.Env) > 0 {
		plan.Add(convertplan.Unsupported, "", subject, "env "+list(def.Env)+": Turborepo keys the task's cache on these variables' values and smith's task cache does not, so changing one does not rerun a cached task; set cache = false on the task, or list the values in its smith.toml env table")
	}
	if len(def.PassThroughEnv) > 0 {
		plan.Add(convertplan.Dropped, "", subject, "passThroughEnv "+list(def.PassThroughEnv)+": smith runs every task with the whole environment, there is no strict mode to pass variables through")
	}
	if def.Persistent != nil && *def.Persistent {
		plan.Add(convertplan.Changed, "", subject, "persistent: smith has no persistent flag; the task stays uncached and smith run keeps it running until it exits or you stop it (smith dev starts several apps' start tasks together)")
	}
	if def.Interactive != nil && *def.Interactive {
		plan.Add(convertplan.Unsupported, "", subject, "interactive: smith run does not forward stdin to a task; run it from its package with npm run")
	}
	if len(def.With) > 0 {
		plan.Add(convertplan.Unsupported, "", subject, "with "+list(def.With)+": Turborepo starts these beside the persistent task; smith run has no such link, so start them with smith dev or list them in the other task's depends_on if they finish")
	}
	if def.OutputLogs != "" || def.OutputMode != "" {
		plan.Add(convertplan.Dropped, "", subject, "outputLogs only sets how much of the task's output turbo prints")
	}
	for _, k := range def.Keys {
		if !knownTaskKeys[k] {
			plan.Add(convertplan.Unsupported, "", subject, "the option "+k+" has no smith equivalent")
		}
	}
}

func list(items []string) string {
	quoted := make([]string, len(items))
	for i, v := range items {
		quoted[i] = `"` + v + `"`
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
