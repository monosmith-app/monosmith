package turboplan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

// TestedTurboMajor is the Turborepo major version the conversion is tested
// against.
const TestedTurboMajor = "2"

// onlyTuningTurbo are turbo.json options that set how turbo itself runs.
var onlyTuningTurbo = map[string]bool{
	"ui": true, "daemon": true, "envMode": true, "cacheDir": true, "concurrency": true, "noUpdateNotifier": true,
	"dangerouslyDisablePackageManagerCheck": true, "futureFlags": true, "experimentalUI": true,
}

// handledSettings are the turbo.json options planned elsewhere.
var handledSettings = map[string]bool{
	"tasks": true, "pipeline": true, "globalDependencies": true, "globalEnv": true,
	"globalPassThroughEnv": true, "remoteCache": true, "$schema": true, "extends": true,
}

// globalInputs are globalDependencies as the inputs of every cached task.
func globalInputs(w *turboworkspace.Workspace) []string {
	out := make([]string, 0, len(w.Turbo.GlobalDependencies))
	for _, g := range w.Turbo.GlobalDependencies {
		out = append(out, RootGlob(g))
	}
	return out
}

// planGlobalSettings reports the root turbo.json options that are not tasks.
func planGlobalSettings(plan *convertplan.Plan, w *turboworkspace.Workspace) {
	t := w.Turbo
	if major := strings.TrimLeft(strings.SplitN(w.TurboVersion, ".", 2)[0], "^~>=< "); major != "" && major != TestedTurboMajor {
		plan.Add(convertplan.Changed, "", "turbo", fmt.Sprintf("the workspace uses Turborepo %s; convert is tested with Turborepo %s, so check this report closely", w.TurboVersion, TestedTurboMajor))
	}
	if len(t.GlobalDependencies) > 0 {
		plan.Add(convertplan.Mapped, "", "globalDependencies", "become inputs of every cached task: "+list(globalInputs(w)))
	}
	if len(t.GlobalEnv) > 0 {
		plan.Add(convertplan.Unsupported, "", "globalEnv", list(t.GlobalEnv)+": Turborepo keys every task's cache on these variables' values and smith's task cache does not; set cache = false on the tasks they affect, or list the values in their smith.toml env table")
	}
	if len(t.GlobalPassThroughEnv) > 0 {
		plan.Add(convertplan.Dropped, "", "globalPassThroughEnv", list(t.GlobalPassThroughEnv)+": smith runs every task with the whole environment, there is no strict mode to pass variables through")
	}
	if len(t.RemoteCache) > 0 {
		plan.Add(convertplan.Unsupported, "", "remoteCache", "Turborepo's (Vercel) remote cache is not carried over; smith caches in .smith/cache and shares it through the bucket or URL you set as [cache] remote in smith.toml (see Tasks and cache)")
	}
	var tuning []string
	for _, k := range t.Keys {
		switch {
		case handledSettings[k]:
		case onlyTuningTurbo[k]:
			tuning = append(tuning, k)
		default:
			plan.Add(convertplan.Unsupported, "", w.Turbo.File, "the option "+k+" has no smith equivalent")
		}
	}
	if len(tuning) > 0 {
		plan.Add(convertplan.Dropped, "", w.Turbo.File, "options that only set how turbo runs: "+strings.Join(tuning, ", "))
	}
}

// reportDefinitions says what each task definition holds that smith has no
// place for, once per definition, and which definitions no package uses.
func reportDefinitions(plan *convertplan.Plan, w *turboworkspace.Workspace) {
	tasks := w.Turbo.TaskSet()
	keys := make([]string, 0, len(tasks))
	for k := range tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		def := tasks[key]
		reportTaskDefinition(plan, "task "+key, def)
		pkg, task, scoped := strings.Cut(key, "#")
		if !scoped {
			pkg, task = "", key
		}
		switch {
		case pkg == "//":
			plan.Add(convertplan.Changed, "", "task "+key, "a root task: smith runs projects, not the workspace root; its script stays in the root package.json, run it with npm run "+task)
		case pkg != "" && w.PackageByName(pkg) == nil:
			plan.Add(convertplan.Dropped, "", "task "+key, pkg+" is not a package of the workspace")
		case !anyPackageHasScript(w, pkg, task):
			plan.Add(convertplan.Dropped, "", "task "+key, "no package has a script named "+task+", so Turborepo had nothing to run")
		}
	}
	for _, p := range w.Packages {
		if p.Turbo == nil {
			continue
		}
		for name, def := range p.Turbo.TaskSet() {
			reportTaskDefinition(plan, p.Dir+"/turbo.json task "+name, def)
		}
	}
}

func anyPackageHasScript(w *turboworkspace.Workspace, pkg, task string) bool {
	for _, p := range w.Packages {
		if pkg != "" && p.Name != pkg {
			continue
		}
		if hasScript(p, task) {
			return true
		}
	}
	return false
}
