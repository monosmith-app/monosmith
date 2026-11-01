package convertplan

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestedNxMajor is the Nx major version the conversion is tested against.
const TestedNxMajor = "23"

// Input is what a plan needs besides the Nx workspace.
type Input struct {
	// DirName is the workspace directory's name, the fallback workspace name.
	DirName string
	// Files are the workspace's files, workspace-relative.
	Files []string
	// Tags are the repository's git tags.
	Tags []string
	// Read returns a workspace-relative file's content, "" when missing.
	Read func(rel string) string
}

var goModule = regexp.MustCompile(`(?m)^module\s+"?([^\s"]+)"?\s*$`)

// Removal plans what a conversion removes of the tool it converts from: its
// configuration files, dependencies and the code that imports it.
type Removal func(plan *Plan, w *nxworkspace.Workspace, files []string, read func(string) string)

// MakePlan decides the whole conversion of an Nx workspace without touching
// the disk.
func MakePlan(w *nxworkspace.Workspace, in Input) *Plan {
	plan := MakePlanWith(w, in, planNxRemoval)
	plan.Source = SourceNx
	if plan.Legacy {
		plan.Source = SourceMnci
	}
	return plan
}

// MakePlanWith is MakePlan for any source workspace described as an Nx one:
// its projects, release set and workspace settings become the plan, and
// removal decides what goes of the tool it came from.
func MakePlanWith(w *nxworkspace.Workspace, in Input, removal Removal) *Plan {
	plan := &Plan{Root: w.Root, Legacy: w.Legacy(), RootScripts: map[string]string{}}
	has := map[string]bool{}
	for _, f := range in.Files {
		has[f] = true
	}
	if major := strings.TrimLeft(strings.SplitN(w.NxVersion, ".", 2)[0], "^~>=< "); major != "" && major != TestedNxMajor {
		plan.Add(Changed, "", "nx", fmt.Sprintf("the workspace uses Nx %s; convert is tested with Nx %s, so check this report closely", w.NxVersion, TestedNxMajor))
	}

	modules := map[string]string{}
	for _, p := range w.Projects {
		if m := goModule.FindStringSubmatch(in.Read(p.Root + "/go.mod")); m != nil {
			modules[p.Root] = m[1]
		}
	}
	plan.Config = WorkspaceSettings(w, in.DirName, func(f string) bool { return has[f] }, modules)
	if plan.Legacy && hasReleasableApp(w) {
		plan.Config.Release.GitHubReleases = plan.Config.Release.GitHubReleases || contains(plan.Config.CI.Providers, "github")
	}
	planCISettings(plan, in, has)

	released := nxworkspace.ReleaseSet(w)
	if plan.Legacy {
		// The old mnci released Go libraries by their <dir>/vX.Y.Z tags,
		// outside nx release.
		for _, p := range w.Projects {
			if p.HasTag("type:go-lib") {
				released[p.Name] = true
			}
		}
	}

	standalone := map[string]bool{}
	var standaloneRoots, keptRoots []string
	for _, p := range w.Projects {
		if Standalone(w, p, released[p.Name]) {
			standalone[p.Name] = true
			standaloneRoots = append(standaloneRoots, p.Root)
			plan.Add(Mapped, path.Base(p.Root), p.Root, "outside the root package.json \"workspaces\" and known to Nx only by its package.json: left out of the workspace with [workspace] exclude, as npm leaves it out; add it to \"workspaces\" instead to make it a project")
			continue
		}
		keptRoots = append(keptRoots, p.Root)
	}
	plan.Config.Workspace.Exclude = ExcludeGlobs(standaloneRoots, keptRoots)

	names := map[string]string{}
	byName := map[string]string{}
	for _, p := range w.Projects {
		if standalone[p.Name] {
			continue
		}
		name := path.Base(p.Root)
		if manifests := manifestsAt(p); len(manifests) > 1 {
			plan.Add(Blocker, name, p.Root, fmt.Sprintf("%s has %s: smith makes one project of each manifest, so they would all be called %q. Move one of them into a folder of its own first", p.Root, strings.Join(manifests, " and "), name))
		}
		if other, dup := byName[name]; dup {
			plan.Add(Blocker, name, p.Root, fmt.Sprintf("%s and %s would both be called %q: smith names projects after their directory, so rename one directory first", other, p.Root, name))
		}
		byName[name] = p.Root
		names[p.Name] = name
	}
	plan.names = names

	pattern := w.Nx.Release.TagPattern()
	for _, p := range w.Projects {
		if !standalone[p.Name] {
			planProject(plan, w, p, names[p.Name], released[p.Name], pattern, in)
		}
	}
	applyTagFormat(plan)
	planRoot(plan, w, names)
	if plan.Legacy {
		for _, f := range legacyTools {
			if has[f] {
				plan.DeletePath(f)
				plan.Add(Dropped, "", f, "the old mnci's Nx release hook; smith release replaces it")
			}
		}
	}
	removal(plan, w, in.Files, in.Read)
	planBuildOutputs(plan, w)
	planPythonTools(plan, in)
	planNodeStack(plan)
	planTaskGroups(plan, w)
	for _, ci := range []string{".github/workflows/ci.yml", "azure-pipelines.yml"} {
		if has[ci] && !strings.Contains(in.Read(ci), "mnci:slot") && !plan.Decided(ci) {
			plan.Decide("", ci, "replaced by the pipeline smith generates; the old one stays in git history, so move any custom steps into its smith:slot sections")
		}
	}
	if plan.Legacy {
		plan.Decide("", "versions", "before 1.0 a feat: commit releases a minor version (the old mnci released a patch); breaking changes still release a minor until 1.0")
	}
	sort.Slice(plan.Projects, func(i, j int) bool { return plan.Projects[i].Name < plan.Projects[j].Name })
	return plan
}

func planProject(plan *Plan, w *nxworkspace.Workspace, p *nxworkspace.Project, name string, released bool, pattern string, in Input) {
	lang, label := Language(p)
	// A Go package of a module declared above it has no manifest of its own:
	// smith.toml declares it as a project (ADR 0052).
	moduleDir, declared := "", false
	if lang == "" {
		if dir, ok := GoModuleAbove(p, in.Files); ok {
			moduleDir, declared, lang = dir, true, "go"
		}
	}
	pp := &ProjectPlan{Nx: p.Name, Name: name, Root: p.Root, Language: lang, Supported: lang != "", Declared: declared}
	plan.Projects = append(plan.Projects, pp)
	if !pp.Supported {
		plan.Add(Unsupported, name, p.Root, label+" project: smith has no plugin for it, so nothing builds, tests or releases it after converting; its project.json is kept for reference")
		return
	}
	kind := Kind(p, released)
	if declared && kind == protocol.ProjectPackage {
		// A module has one version, so a package of one cannot be released alone.
		kind = protocol.ProjectLib
		plan.Add(Changed, name, "kind", "a lib: a package of a module cannot be released on its own, since the module has one version")
	}
	pp.Kind = kind
	if declared {
		o := plan.Override(name)
		o.Root, o.Language, o.Kind = p.Root, "go", string(kind)
		plan.SetOverride(name, o)
		plan.Add(Mapped, name, p.Root, "declared in smith.toml as a Go project of the module at "+moduleDir+": it has no go.mod of its own, so smith builds and tests it from the module's")
	}
	pp.KindMarker = !declared && NeedsKindMarker(p, lang, kind)
	if pp.KindMarker {
		plan.Add(Mapped, name, "kind", fmt.Sprintf("%s is declared in its manifest, since %s/ does not imply it", kind, strings.SplitN(p.Root, "/", 2)[0]))
	}
	appReleases := false
	if kind == protocol.ProjectApp && released {
		if lang == "go" {
			o := plan.Override(name)
			o.Release = AppRelease(p, p.Name)
			plan.SetOverride(name, o)
			appReleases = true
			plan.Add(Mapped, name, "release", "releases with per-platform binaries named "+o.Release.AssetName)
		} else {
			plan.Add(Unsupported, name, "release", "releasing a "+label+" app is not supported yet; it converts as an app that does not release")
		}
	}
	if p.HasFile(Dockerfile) {
		planImage(plan, name, kind)
	}
	if p.HasTag("type:vscode-extension") {
		plan.Add(Unsupported, name, "release", "publishing a VS Code extension is not supported yet")
	}
	if kind == protocol.ProjectPackage || appReleases {
		def := name + "@"
		if lang == "go" && kind == protocol.ProjectPackage {
			def = p.Root + "/v"
		}
		prefix := def
		// The old mnci tagged Go libraries <dir>/vX.Y.Z, which is smith's
		// default too; anything else continues the Nx pattern's tags.
		legacyGoLib := plan.Legacy && lang == "go" && kind == protocol.ProjectPackage
		if !legacyGoLib {
			var err error
			if prefix, err = TagPrefix(pattern, p, def, in.Tags); err != nil {
				plan.Add(Manual, name, "release tags", err.Error())
			}
		}
		if lang != "go" || kind != protocol.ProjectPackage { // a Go module's own prefix is no choice
			_, taggedUnderDefault := newestUnder(in.Tags, def)
			plan.tagChoices = append(plan.tagChoices, tagChoice{project: name, prefix: prefix, fresh: prefix == def && !taggedUnderDefault})
		}
		if prefix != def {
			o := plan.Override(name)
			o.TagPrefix = prefix
			plan.SetOverride(name, o)
			plan.Add(Mapped, name, "release tags", "continues its tags "+prefix+"<version>")
		}
	}
	legacy := map[string]string{}
	if plan.Legacy {
		legacy = legacyTargets
	}
	mapTargets(plan, w, p, pp, legacy)
}

// planImage turns on the docker convention for an app with a Dockerfile, so
// it gets an image task and its release ships the image.
func planImage(plan *Plan, name string, kind protocol.ProjectKind) {
	if kind != protocol.ProjectApp {
		plan.Add(Changed, name, Dockerfile, "kept, but the docker convention builds images of apps only, and this converts as a "+string(kind))
		return
	}
	if plan.Config.Conventions == nil {
		plan.Config.Conventions = map[string]workspace.ConventionConfig{}
	}
	plan.Config.Conventions["docker"] = workspace.ConventionConfig{}
	plan.Add(Changed, name, Dockerfile, "[conventions.docker] gives the app an image task (docker build after its build), and its release pushes the image to [conventions.docker] registry once you set one")
}

// planRoot converts the root project's targets and the root scripts.
func planRoot(plan *Plan, w *nxworkspace.Workspace, names map[string]string) {
	scripts := w.RootPackage.Scripts
	for name, t := range w.RootTargets {
		if t.Executor == "nx:run-script" {
			continue
		}
		if reason, ok := plumbing[name]; ok {
			plan.Add(Dropped, "", "root target "+name, reason)
			continue
		}
		if reason, ok := plumbing[t.Executor]; ok {
			plan.Add(Dropped, "", "root target "+name, reason)
			continue
		}
		line, cwd, _, err := commandOf(t)
		if err != nil || cwd != "." {
			plan.Add(Manual, "", "root target "+name, "the workspace root's target is not converted; smith runs projects only")
			continue
		}
		script := name
		if _, taken := scripts[script]; taken {
			script += ":workspace"
		}
		plan.RootScripts[script] = line
		plan.Add(Changed, "", "root target "+name, fmt.Sprintf("becomes the root script %q (smith runs projects, not the workspace root): %s", script, line))
	}
	removedProjectScripts := 0
	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cmd := scripts[k]
		if plan.Legacy {
			if reason, ok := legacyRootScripts[k]; ok {
				plan.RootScripts[k] = ""
				plan.Add(Dropped, "", "script "+k, reason)
				continue
			}
			if legacyProjectScript(k, names) {
				plan.RootScripts[k] = ""
				removedProjectScripts++
				continue
			}
		}
		if !mentionsNx(cmd) {
			if tools := MissingScriptTools(plan.Root, cmd); len(tools) > 0 {
				plan.Add(Manual, "", "script "+k, "starts "+strings.Join(tools, ", ")+", which is not installed, and convert keeps the script as it is: install the tool (npm i -D "+strings.Join(tools, " ")+" at the workspace root) or remove the script")
			}
			continue
		}
		if rewritten, ok := RewriteScript(cmd, names); ok {
			plan.RootScripts[k] = rewritten
			plan.Add(Mapped, "", "script "+k, rewritten)
		} else if gap := ScriptGap(cmd); gap != "" {
			plan.Add(Manual, "", "script "+k, gap+": "+cmd)
		} else {
			plan.Add(Manual, "", "script "+k, "runs Nx in a way smith cannot translate: "+cmd)
		}
	}
	if removedProjectScripts > 0 {
		plan.Add(Dropped, "", "scripts", fmt.Sprintf("%d per-project scripts (<project>:test and the like); use smith run <task> -p <project>", removedProjectScripts))
	}
}

// planNodeStack makes node-jest the workspace's default Node stack, so a
// Node project that records no stack, and any added later, gets the lint and
// test tasks. A script a project already has still wins over the stack's.
func planNodeStack(plan *Plan) {
	for _, p := range plan.Projects {
		if p.Language != "node" {
			continue
		}
		if plan.Config.DefaultStacks == nil {
			plan.Config.DefaultStacks = map[string]string{}
		}
		plan.Config.DefaultStacks["node"] = "node-jest"
		plan.Add(Mapped, "", "[default_stacks]", `node = "node-jest": Node projects that record no stack, and new ones, get ESLint and Jest tasks; a script a project already has wins`)
		return
	}
}

// planPythonTools makes sure an existing requirements-dev.txt holds every
// tool the Python plugin runs.
func planPythonTools(plan *Plan, in Input) {
	python := false
	for _, p := range plan.Projects {
		python = python || p.Language == "python"
	}
	current := in.Read(venv.RequirementsFile)
	if !python || current == "" {
		return
	}
	have := map[string]bool{}
	for _, line := range strings.Split(current, "\n") {
		have[requirementName(line)] = true
	}
	var missing []string
	for _, line := range strings.Split(venv.Requirements, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !have[requirementName(line)] {
			missing = append(missing, line)
		}
	}
	if len(missing) == 0 {
		return
	}
	plan.WriteFile(venv.RequirementsFile, strings.TrimRight(current, "\n")+"\n"+strings.Join(missing, "\n")+"\n")
	plan.Add(Changed, "", venv.RequirementsFile, "adds the tools the Python plugin runs: "+strings.Join(missing, ", "))
}

func legacyProjectScript(script string, names map[string]string) bool {
	for nx, name := range names {
		for _, suffix := range legacyProjectScriptSuffixes {
			if script == name+suffix || script == nx+suffix {
				return true
			}
		}
	}
	return false
}

func mentionsNx(cmd string) bool {
	for _, w := range strings.Fields(cmd) {
		if w == "nx" || w == "mnci" {
			return true
		}
	}
	return false
}

func hasReleasableApp(w *nxworkspace.Workspace) bool {
	for _, p := range w.Projects {
		if p.HasTag("release:go") {
			return true
		}
	}
	return false
}

// requirementName is the lower-case distribution name of a requirements line.
func requirementName(line string) string {
	line = strings.TrimSpace(line)
	if i := strings.IndexAny(line, "=<>~!;[# "); i >= 0 {
		line = line[:i]
	}
	return strings.ToLower(line)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
