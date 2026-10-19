package turboplan

import (
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

// MakePlan decides the whole conversion of a Turborepo workspace without
// touching the disk. The projects, kinds, release settings and pipelines are
// planned as for an Nx workspace; this adds turbo.json's tasks, the scripts
// that call turbo, pnpm or yarn, and the move to npm workspaces.
func MakePlan(w *turboworkspace.Workspace, in convertplan.Input) *convertplan.Plan {
	plan := convertplan.MakePlanWith(w.NxShape(), in, func(plan *convertplan.Plan, _ *nxworkspace.Workspace, files []string, read func(string) string) {
		planTurboRemoval(plan, w, files, read)
	})
	plan.Source = convertplan.SourceTurborepo

	names := projectNames(plan, w)
	c := taskContext{plan: plan, w: w, names: names, global: globalInputs(w)}
	scripts := 0
	for _, p := range w.Packages {
		pp := plan.ProjectAt(p.Dir)
		if pp == nil || !pp.Supported {
			continue
		}
		scripts += len(p.JSON.Scripts)
		mapPackageTasks(c, p, pp)
		rewritePackageScripts(plan, p, pp, names)
	}
	if scripts > 0 {
		plan.Add(convertplan.Mapped, "", "scripts", "package.json scripts stay scripts: smith's Node plugin runs each as a task (npm run <script>), so smith run build runs every package's build script")
	}
	rewriteRootScripts(plan, w, names)
	reportDefinitions(plan, w)
	planGlobalSettings(plan, w)
	return plan
}

// projectNames maps what a script can call a package (its package.json name,
// its directory's name, its name without the scope) to the smith project.
func projectNames(plan *convertplan.Plan, w *turboworkspace.Workspace) map[string]string {
	names := map[string]string{}
	bare := map[string][]string{}
	for _, p := range w.Packages {
		pp := plan.ProjectAt(p.Dir)
		if pp == nil {
			continue
		}
		names[p.Name] = pp.Name
		names[path.Base(p.Dir)] = pp.Name
		if scope, name, scoped := strings.Cut(p.Name, "/"); scoped && strings.HasPrefix(scope, "@") {
			bare[name] = append(bare[name], pp.Name)
		}
	}
	for name, projects := range bare {
		if _, taken := names[name]; !taken && len(projects) == 1 {
			names[name] = projects[0]
		}
	}
	return names
}

// rewritePackageScripts translates the scripts of one package that call turbo,
// pnpm or yarn.
func rewritePackageScripts(plan *convertplan.Plan, p *turboworkspace.Package, pp *convertplan.ProjectPlan, names map[string]string) {
	keys := make([]string, 0, len(p.JSON.Scripts))
	for k := range p.JSON.Scripts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		r := RewriteScript(p.JSON.Scripts[k], names)
		switch {
		case !r.Changed:
		case r.Problem != "":
			plan.Add(convertplan.Manual, pp.Name, "script "+k, r.Problem+": "+p.JSON.Scripts[k])
		default:
			if pp.Scripts == nil {
				pp.Scripts = map[string]string{}
			}
			pp.Scripts[k] = r.Script
			plan.Add(convertplan.Mapped, pp.Name, "script "+k, r.Script)
			reportDroppedFlags(plan, pp.Name, "script "+k, r.Dropped)
		}
	}
}

// rewriteRootScripts translates the root package.json scripts. A script that
// only keeps pnpm or yarn as the package manager (npx only-allow pnpm) goes.
func rewriteRootScripts(plan *convertplan.Plan, w *turboworkspace.Workspace, names map[string]string) {
	keys := make([]string, 0, len(w.RootPackage.Scripts))
	for k := range w.RootPackage.Scripts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cmd := w.RootPackage.Scripts[k]
		if strings.Contains(cmd, "only-allow") {
			plan.RootScripts[k] = ""
			plan.Add(convertplan.Dropped, "", "script "+k, "it only enforced a package manager ("+cmd+"); smith install uses npm")
			continue
		}
		r := RewriteScript(cmd, names)
		switch {
		case !r.Changed:
		case r.Problem != "":
			plan.Add(convertplan.Manual, "", "script "+k, r.Problem+": "+cmd)
		default:
			plan.RootScripts[k] = r.Script
			plan.Add(convertplan.Mapped, "", "script "+k, r.Script)
			reportDroppedFlags(plan, "", "script "+k, r.Dropped)
		}
	}
}

func reportDroppedFlags(plan *convertplan.Plan, project, subject string, flags []string) {
	if len(flags) > 0 {
		plan.Add(convertplan.Dropped, project, subject, "turbo flags that only tune turbo's output: "+strings.Join(flags, ", "))
	}
}
