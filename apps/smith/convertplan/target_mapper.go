package convertplan

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// pluginTargets are executors whose job the smith language plugin does.
var pluginTargets = map[string]string{
	"@nx-go/nx-go:build":            "go",
	"@nx-go/nx-go:test":             "go",
	"@nx-go/nx-go:lint":             "go",
	"@mnci/nx-python-pip:lint":      "python",
	"@mnci/nx-python-pip:test":      "python",
	"@mnci/nx-python-pip:build":     "python",
	"@mnci/nx-python-pip:typecheck": "python",
}

// plumbing are targets and executors that only serve Nx itself.
var plumbing = map[string]string{
	"build-deps":                    "Nx's build-deps helper; smith builds dependencies first on its own",
	"watch-deps":                    "Nx's watch helper",
	"nx-release-publish":            "smith release publishes",
	"nx:noop":                       "an Nx no-op",
	"@nx/js:release-publish":        "smith release publishes",
	"@mnci/nx-python-pip:publish":   "smith release publishes",
	"@nx/js:prune-lockfile":         "an Nx deployment helper with no smith equivalent",
	"@nx/js:copy-workspace-modules": "an Nx deployment helper with no smith equivalent",
	"@nx/js:verdaccio":              "a local registry helper",
	"@nx-go/nx-go:serve":            "smith run start runs the app (go run .)",
	"@nx-go/nx-go:tidy":             "run go mod tidy directly",
	"@nx-go/nx-go:generate":         "run go generate ./... directly",
}

// packageTask is the task smith ci pack runs into dist/drop.
const packageTask = "package"

// pluginPackages are the languages whose plugin gives every app a package
// task, and what it makes. The old mnci's package target of such an app
// zipped the same thing, so the plugin's task replaces it; any other package
// target is converted like any target.
var pluginPackages = map[string]string{
	"go":     "binary",
	"python": "wheel",
}

// mapTargets converts one project's targets: each either stays with the
// language plugin, becomes a package.json script (Node), becomes an
// smith.toml task, or is reported.
func mapTargets(plan *Plan, w *nxworkspace.Workspace, p *nxworkspace.Project, pp *ProjectPlan, legacy map[string]string) {
	names := make([]string, 0, len(p.Targets))
	for name := range p.Targets {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t := p.Targets[name]
		if reason, ok := legacy[name]; ok {
			plan.Add(Dropped, pp.Name, name, reason)
			continue
		}
		if name == packageTask && plan.Legacy {
			if made, ok := pluginPackages[pp.Language]; ok {
				plan.Add(Changed, pp.Name, name, "the "+pp.Language+" plugin's package task replaces the old mnci's: smith ci pack puts the app's "+made+" in dist/drop")
				continue
			}
		}
		if strings.Contains(name, "-ci--") {
			plan.Add(Dropped, pp.Name, name, "Nx's per-file split of a test target for CI")
			continue
		}
		if reason, ok := plumbing[name]; ok {
			plan.Add(Dropped, pp.Name, name, reason)
			continue
		}
		if reason, ok := plumbing[t.Executor]; ok {
			severity := Dropped
			if strings.Contains(reason, "no smith equivalent") {
				severity = Changed
			}
			plan.Add(severity, pp.Name, name, reason)
			continue
		}
		switch t.Executor {
		case "nx:run-script":
			continue // an existing package.json script: the Node plugin runs it
		}
		if lang, ok := pluginTargets[t.Executor]; ok && lang == pp.Language {
			pluginTarget(plan, pp, name, t)
			continue
		}
		if t.Executor == "nx:run-commands" || t.Executor == "@nx/workspace:run-commands" || t.Executor == "@nrwl/workspace:run-commands" || (t.Executor == "" && t.Command != "") {
			line, cwd, env, err := commandOf(t)
			if err != nil {
				plan.Add(Manual, pp.Name, name, "not converted: "+err.Error()+"; add it to smith.toml by hand")
				continue
			}
			commandTarget(plan, pp, name, t, line, cwd, env)
			continue
		}
		if why := executorGap(t); why != "" {
			plan.Add(Manual, pp.Name, name, why)
			continue
		}
		if line, needs, ok := executorCommand(w, p, t); ok && pp.Language == "node" {
			for _, n := range needs {
				if plan.AddDevDependencies == nil {
					plan.AddDevDependencies = map[string]string{}
				}
				plan.AddDevDependencies[n] = ""
			}
			script := name
			if t.Executor == "@nx/js:node" {
				if _, taken := p.Targets["start"]; !taken {
					script = "start"
				}
			}
			copied := false
			if copiesAssets(p, t) {
				var out string
				t.Option("outputPath", &out)
				if copy, ok := AssetCopy(p.Root, t, relativeTo(p.Root, out)); ok && out != "" {
					line += " && " + copy
					copied = true
				} else {
					plan.Add(Manual, pp.Name, name, "the build copied assets ("+t.Executor+" assets option); add the copy to the build script")
				}
			}
			setScript(plan, pp, script, line, t, nil)
			plan.Add(Changed, pp.Name, name, fmt.Sprintf("%s becomes the script %q: %s", t.Executor, script, line))
			if t.Executor == "@nx/esbuild:esbuild" {
				opts := MapEsbuildOptions(t)
				unmapped := opts.Unmapped
				if !esbuildBundles(t) {
					plan.Add(Changed, pp.Name, name, "bundle: false: like Nx, it transpiles each file the tsconfig includes into the same place under the output, bundling nothing")
					// The transpile script takes none of them.
					if len(opts.Flags) > 0 || opts.Sourcemap != "--sourcemap" {
						unmapped = append(unmapped, "esbuildOptions (with bundle: false)")
					}
				}
				if len(unmapped) > 0 {
					plan.Add(Manual, pp.Name, name, "esbuildOptions not carried to the build script: "+strings.Join(unmapped, ", ")+"; add the matching esbuild flags to the "+script+" script")
				}
			}
			if copied {
				plan.Add(Changed, pp.Name, name, "its assets are copied after the build by a node one-liner, as Nx copied them (fs.globSync: Node 22 or later)")
			}
			continue
		}
		plan.Add(Manual, pp.Name, name, fmt.Sprintf("the executor %s has no smith equivalent; add the command it runs to smith.toml by hand", t.Executor))
	}
}

// pluginTarget keeps a target with the language plugin, carrying over what
// the plugin's default does not already do.
func pluginTarget(plan *Plan, pp *ProjectPlan, name string, t nxworkspace.Target) {
	o := workspace.TaskOverride{DependsOn: extraDependsOn(plan, pp, name, t)}
	if t.Executor == "@nx-go/nx-go:lint" {
		var linter string
		if t.Option("linter", &linter) && linter != "" && linter != "go vet" {
			var args []string
			t.Option("args", &args)
			o.Command = append(strings.Fields(linter), args...)
			if t.Parallelism != nil && !*t.Parallelism && o.Command[0] == "golangci-lint" {
				// Nx ran it one at a time; golangci-lint waits for its own
				// lock instead of failing when smith runs projects in parallel.
				o.Command = append(o.Command, "--allow-serial-runners")
			}
		}
	} else if t.Parallelism != nil && !*t.Parallelism {
		plan.Add(Changed, pp.Name, name, "Nx ran it one at a time (parallelism: false); smith may run it alongside other tasks")
	}
	if o.Command == nil && o.DependsOn == nil {
		plan.Add(Mapped, pp.Name, name, "run by the "+pp.Language+" plugin")
		return
	}
	plan.SetTask(pp.Name, name, o)
	plan.Add(Mapped, pp.Name, name, "run by the "+pp.Language+" plugin, with its settings kept in smith.toml")
}

// commandTarget converts one shell command. A Node project runs it as a
// package.json script when it runs in the project root (npm runs scripts
// through a shell, so nothing is lost); anything else becomes a smith.toml
// task with an argv.
func commandTarget(plan *Plan, pp *ProjectPlan, name string, t nxworkspace.Target, line, cwd string, env map[string]string) {
	if pp.Language == "node" && cwd == pp.Root {
		setScript(plan, pp, name, line, t, env)
		plan.Add(Mapped, pp.Name, name, "package.json script: "+line)
		return
	}
	cl, err := SplitCommandLine(line)
	if err != nil {
		if pp.Language == "node" {
			up := relativeTo(pp.Root, cwd)
			setScript(plan, pp, name, "cd "+up+" && "+line, t, env)
			plan.Add(Mapped, pp.Name, name, "package.json script (it needs a shell): cd "+up+" && "+line)
			return
		}
		plan.Add(Manual, pp.Name, name, "not converted: "+err.Error()+"; split it into separate tasks or move it to a script")
		return
	}
	for k, v := range env {
		if cl.Env == nil {
			cl.Env = map[string]string{}
		}
		cl.Env[k] = v
	}
	if cwd == pp.Root {
		cwd = ""
	}
	if pp.Language == "python" {
		// smith puts the workspace's .venv first on PATH, and a Windows venv
		// has python but no python3.
		for i, a := range cl.Argv {
			if a == "python3" {
				cl.Argv[i] = "python"
			}
		}
	}
	customTask(plan, pp, name, t, cl.Argv, cwd)
	if len(cl.Env) > 0 {
		o := plan.Override(pp.Name)
		task := o.Tasks[name]
		task.Env = cl.Env
		plan.SetTask(pp.Name, name, task)
	}
}

// customTask writes a smith.toml task with everything Nx knew about it.
func customTask(plan *Plan, pp *ProjectPlan, name string, t nxworkspace.Target, argv []string, cwd string) {
	o := workspace.TaskOverride{Command: argv, Cwd: cwd, DependsOn: dependsOn(plan, pp, name, t), Inputs: inputs(pp.Root, t), Outputs: outputs(pp.Root, t)}
	cache := t.Cache != nil && *t.Cache && !t.Continuous
	if cache {
		o.Cache = &cache
	}
	plan.SetTask(pp.Name, name, o)
	plan.Add(Mapped, pp.Name, name, "smith.toml task: "+strings.Join(argv, " "))
}

// setScript records a package.json script and the task settings npm cannot
// carry: environment variables and unusual dependencies.
func setScript(plan *Plan, pp *ProjectPlan, name, line string, t nxworkspace.Target, env map[string]string) {
	if pp.Scripts == nil {
		pp.Scripts = map[string]string{}
	}
	pp.Scripts[name] = line
	o := workspace.TaskOverride{DependsOn: extraDependsOn(plan, pp, name, t)}
	if len(env) > 0 {
		o.Env = env
	}
	if o.DependsOn != nil || o.Env != nil {
		plan.SetTask(pp.Name, name, o)
	}
}

// dependsOn translates Nx dependsOn. A target of another project
// ({projects, target} or "project:target") becomes "<smith name>:<target>";
// a project smith cannot name (a pattern, a tag, an unknown project) is
// reported, since leaving it out changes the order tasks run in.
func dependsOn(plan *Plan, pp *ProjectPlan, task string, t nxworkspace.Target) []string {
	var out, others []string
	add := func(dep string) {
		for _, have := range out {
			if have == dep {
				return
			}
		}
		out = append(out, dep)
	}
	other := func(dep string, cross bool) {
		add(dep)
		if cross && !contains(others, dep) {
			others = append(others, dep)
		}
	}
	defer func() {
		if len(others) > 0 {
			plan.Add(Mapped, pp.Name, task, "runs after "+strings.Join(others, ", ")+", as Nx did (depends_on)")
		}
	}()
	for _, d := range t.DependsOn {
		switch {
		case d.Target == "":
		case len(d.Projects) > 0:
			for _, project := range d.Projects {
				if dep, cross, ok := projectTarget(plan, pp, project, d.Target); ok {
					other(dep, cross)
				} else {
					plan.Add(Manual, pp.Name, task, fmt.Sprintf("Nx ran %s's %s first; smith cannot name that project, so add it to depends_on by hand (\"<project>:%s\")", project, d.Target, d.Target))
				}
			}
		case d.Dependencies:
			add("^" + d.Target)
		default:
			if project, target, found := strings.Cut(d.Target, ":"); found {
				if dep, cross, ok := projectTarget(plan, pp, project, target); ok {
					other(dep, cross)
					continue
				}
			}
			add(d.Target)
		}
	}
	return out
}

// projectTarget names another project's target the way smith's depends_on
// does (<project>:<target>, cross), or the target alone for the project
// itself. ok is false when the Nx project resolves to no smith project.
func projectTarget(plan *Plan, pp *ProjectPlan, project, target string) (dep string, cross, ok bool) {
	if project == "self" || project == pp.Nx {
		return target, false, true
	}
	if project == "dependencies" {
		return "^" + target, false, true
	}
	name, ok := smithName(project, plan.names)
	if !ok {
		return "", false, false
	}
	if name == pp.Name {
		return target, false, true
	}
	return name + ":" + target, true, true
}

// extraDependsOn is dependsOn unless it is only "^build", which the
// plugins' tasks already depend on. Anything else is kept: a typecheck that
// Nx ran after the project's own build must not race that build.
func extraDependsOn(plan *Plan, pp *ProjectPlan, task string, t nxworkspace.Target) []string {
	deps := dependsOn(plan, pp, task, t)
	for _, d := range deps {
		if d != "^build" {
			return deps
		}
	}
	return nil
}

// inputs translates Nx inputs to smith globs. Named inputs (default,
// production) mean the project's files, which is also smith's default, so a
// target whose inputs are all named keeps the default.
func inputs(root string, t nxworkspace.Target) []string {
	var out []string
	named := false
	for _, in := range t.Inputs {
		switch {
		case in.Text == "":
			continue
		case strings.HasPrefix(in.Text, "{projectRoot}/") || strings.HasPrefix(in.Text, "!{projectRoot}/"):
			out = append(out, strings.Replace(in.Text, "{projectRoot}/", "", 1))
		case strings.HasPrefix(in.Text, "{workspaceRoot}/") || strings.HasPrefix(in.Text, "!{workspaceRoot}/"):
			out = append(out, in.Text)
		default:
			named = true
		}
	}
	if named && len(out) > 0 {
		out = append([]string{"**/*"}, out...)
	}
	return out
}

// outputs translates Nx outputs to directories relative to the project
// root; a glob keeps the directory it starts in.
func outputs(root string, t nxworkspace.Target) []string {
	var out []string
	seen := map[string]bool{}
	for _, o := range t.Outputs {
		o = strings.ReplaceAll(o, "\\", "/")
		if strings.Contains(o, "{options.") {
			continue
		}
		o = strings.Replace(o, "{projectRoot}", "{workspaceRoot}/"+root, 1)
		o = strings.TrimPrefix(o, "{workspaceRoot}/")
		if i := strings.IndexAny(o, "*{?["); i >= 0 {
			o = path.Dir(o[:i] + "x")
		}
		if path.Ext(o) != "" {
			o = path.Dir(o)
		}
		r := relativeTo(root, o)
		if r == "." || seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}

// copiesAssets reports whether a build's assets option copies real files
// (not just a .gitkeep).
func copiesAssets(p *nxworkspace.Project, t nxworkspace.Target) bool {
	var assets []json.RawMessage
	if !t.Option("assets", &assets) {
		return false
	}
	for _, raw := range assets {
		var dir string
		if json.Unmarshal(raw, &dir) != nil {
			var obj struct {
				Input string `json:"input"`
			}
			_ = json.Unmarshal(raw, &obj)
			dir = obj.Input
		}
		dir = strings.TrimSuffix(strings.ReplaceAll(dir, "\\", "/"), "/")
		for _, f := range p.Files {
			if strings.HasPrefix(f, dir+"/") && path.Base(f) != ".gitkeep" {
				return true
			}
		}
	}
	return false
}
