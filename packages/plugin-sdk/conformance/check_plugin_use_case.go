package conformance

import (
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options configure Check.
type Options struct {
	// Workspace is the empty directory the plugin was started in. When set, a
	// plugin that generates projects is asked to generate an app, a lib and a
	// package there (only the kinds it declares, when it declares any), which a language plugin must then discover as
	// generated. Generation may install dependencies, so it is opt-in.
	Workspace string
}

// Stack is the stack name generated projects are asked to record.
const Stack = "conformance"

// Check runs every check against a connected plugin.
func Check(c *host.Client, opts Options) Report {
	info := c.Info
	r := Report{Plugin: info.Plugin}
	r.add("names itself", info.Plugin.Name != "" && info.Plugin.Version != "", "initialize returned name %q, version %q; both are required", info.Plugin.Name, info.Plugin.Version)
	r.add("has a known kind", info.Kind == protocol.KindLanguage || info.Kind == protocol.KindConvention, "kind %q is neither language nor convention", info.Kind)
	checkLanguages(info, &r)
	roles := info.EffectiveRoles()
	for _, role := range roles {
		r.add("plays a known role: "+string(role), role.Known(), "roles are %v", protocol.Roles)
	}
	caps := info.Capabilities
	if slices.Contains(roles, protocol.RoleLanguage) {
		r.add("a language plugin discovers projects", caps.Discover && len(info.Manifests) > 0, "it must implement discover and name its manifest files")
	}
	if slices.Contains(roles, protocol.RoleScaffold) {
		r.add("a scaffold generates projects", caps.Generate, "it must implement generate")
	}
	for _, role := range []protocol.Role{protocol.RoleToolchain, protocol.RoleStructure, protocol.RoleDeploy} {
		if slices.Contains(roles, role) && !slices.Contains(roles, protocol.RoleLanguage) {
			r.add("a "+string(role)+" plugin does something", caps.Contribute || caps.Generate || caps.CI || caps.Doctor, "it implements none of contribute, generate, ci and doctor")
		}
	}
	if len(info.Provides) > 0 {
		r.add("only a toolchain provides tasks", slices.Contains(roles, protocol.RoleToolchain), "it provides %v without the toolchain role", info.Provides)
	}
	if info.Kind == protocol.KindConvention && slices.Contains(roles, protocol.RoleToolchain) {
		r.add("a toolchain names the slots it fills", len(info.Provides) > 0, "a toolchain plugin must list the tasks it provides")
	}
	for _, ref := range info.Requires {
		r.add("requires "+ref.String()+": a known reference", validRef(ref), "a reference sets a known level, a name or a slot")
	}
	for _, ref := range info.Conflicts {
		r.add("conflicts with "+ref.String()+": a known reference", validRef(ref) && ref.Name != info.Plugin.Name, "a reference sets a known level, a name or a slot, and is not the plugin itself")
	}
	names := map[string]bool{}
	for _, st := range info.Stacks {
		r.add("stack "+st.Name+" is complete", st.Name != "" && st.Language != "" && !names[st.Name], "a stack needs a unique name and a language plugin")
		names[st.Name] = true
	}

	checkScheduling(c, &r)

	call := func(check, method string, params, result any) bool {
		if !c.Supports(method) {
			return false
		}
		err := c.Call(method, params, result, 20*time.Minute)
		r.add(check, err == nil, "%v", err)
		return err == nil
	}
	call("discover answers with no manifests", protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{}}, &protocol.DiscoverResult{})
	call("contribute answers with no projects", protocol.MethodContribute, protocol.ContributeParams{Projects: []protocol.Project{}}, &protocol.ContributeResult{})
	checkFacts(c, &r)
	call("ci answers with no projects", protocol.MethodCI, protocol.CIParams{Projects: []protocol.Project{}}, &protocol.CIResult{})
	call("doctor answers with no projects", protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{}}, &protocol.DoctorResult{})
	call("derive answers with no projects", protocol.MethodDerive, protocol.DeriveParams{Projects: []protocol.Project{}}, &protocol.DeriveResult{})

	if opts.Workspace != "" && caps.Generate {
		generate(c, opts.Workspace, &r)
	}
	return r
}

// checkLanguages checks the languages a plugin declares are usable names. A
// convention that declares none says in the check's name that it applies to
// every language, so the report shows what smith will do with it.
func checkLanguages(info protocol.InitializeResult, r *Report) {
	if info.Kind == protocol.KindConvention && len(info.Languages) == 0 {
		r.add("declares no languages, so it applies to projects of every language", true, "")
		return
	}
	if len(info.Languages) == 0 {
		return
	}
	var bad []string
	seen := map[string]bool{}
	for _, l := range info.Languages {
		if strings.TrimSpace(l) == "" || l != strings.TrimSpace(l) || seen[l] {
			bad = append(bad, "\""+l+"\"")
		}
		seen[l] = true
	}
	what := "declares the languages it serves"
	if info.Kind == protocol.KindConvention {
		what = "declares the languages it applies to"
	}
	r.add(what+": "+strings.Join(info.Languages, ", "), len(bad) == 0, "languages must be non-empty names, each listed once; problem with %s", strings.Join(bad, ", "))
}

// generate asks for one project of each kind the plugin declares it scaffolds
// (all three when it declares none), checks the files it reports
// exist, and, for a language plugin, that discovery returns what it made.
func generate(c *host.Client, ws string, r *Report) {
	language := slices.Contains(c.Info.EffectiveRoles(), protocol.RoleLanguage)
	var made []protocol.GenerateParams
	for _, kind := range c.Info.ScaffoldKinds() {
		name := "c" + string(kind)
		params := protocol.GenerateParams{
			Kind: kind, Name: name, Directory: projectkind.Directory(kind) + "/" + name, Stack: Stack,
			Workspace: protocol.WorkspaceInfo{Name: "conformance", Scope: "@conformance", ModulePrefix: "example.com/conformance"},
		}
		var res protocol.GenerateResult
		if err := c.Call(protocol.MethodGenerate, params, &res, 20*time.Minute); err != nil {
			r.add("generates "+article(kind), false, "%v", err)
			continue
		}
		var missing []string
		for _, f := range res.Files {
			if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(f))); err != nil {
				missing = append(missing, f)
			}
		}
		r.add("generates "+article(kind), len(res.Files) > 0 && len(missing) == 0, "files reported but not written: %v (of %v)", missing, res.Files)
		made = append(made, params)
	}
	if !language || len(made) == 0 {
		return
	}
	var res protocol.DiscoverResult
	if err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: manifests(ws, c.Info.Manifests)}, &res, 20*time.Minute); err != nil {
		r.add("discovers what it generated", false, "%v", err)
		return
	}
	found := map[string]protocol.Project{}
	for _, p := range res.Projects {
		found[p.Root] = p
	}
	for _, params := range made {
		p, ok := found[params.Directory]
		r.add("discovers the generated "+string(params.Kind), ok, "no project at %s among %d discovered", params.Directory, len(res.Projects))
		if !ok {
			continue
		}
		r.add("the generated "+string(params.Kind)+" keeps its kind", p.Kind == params.Kind, "discovered as %q", p.Kind)
		r.add("the generated "+string(params.Kind)+" records its stack", p.Stack == Stack, "discovered with stack %q; generate must record params.stack in the manifest", p.Stack)
		var empty []string
		for task, def := range p.Tasks {
			if len(def.Command) == 0 {
				empty = append(empty, task)
			}
		}
		r.add("the generated "+string(params.Kind)+"'s tasks have commands", len(empty) == 0, "tasks without a command: %v", empty)
		if c.Supports(protocol.MethodVersion) {
			var v protocol.VersionResult
			err := c.Call(protocol.MethodVersion, protocol.VersionParams{Project: p, Version: "1.2.3", DryRun: true}, &v, time.Minute)
			r.add("writes the "+string(params.Kind)+"'s version (dry run)", err == nil, "%v", err)
		}
		if c.Supports(protocol.MethodRemove) {
			var rm protocol.RemoveResult
			err := c.Call(protocol.MethodRemove, protocol.RemoveParams{Project: p, DryRun: true}, &rm, time.Minute)
			r.add("would remove the "+string(params.Kind)+" (dry run)", err == nil, "%v", err)
		}
		if c.Supports(protocol.MethodChangeKind) {
			checkChangeKind(c, ws, p, res.Projects, r)
		}
		if params.Kind == protocol.ProjectPackage && c.Supports(protocol.MethodPublish) {
			var pub protocol.PublishResult
			err := c.Call(protocol.MethodPublish, protocol.PublishParams{Project: p, Version: "1.2.3", DryRun: true}, &pub, 5*time.Minute)
			r.add("publishes the package (dry run)", err == nil, "%v", err)
		}
		if params.Kind == protocol.ProjectPackage && c.Supports(protocol.MethodPublished) {
			var pub protocol.PublishedResult
			err := c.Call(protocol.MethodPublished, protocol.PublishedParams{Project: p}, &pub, time.Minute)
			r.add("says what the registry has of the package", err == nil, "%v (an unreachable registry is a result, not an error)", err)
		}
	}
}

// kindChangeTarget is the kind a generated project is asked to become: a lib
// or an app is published as a package, a package goes back to a lib.
func kindChangeTarget(kind protocol.ProjectKind) protocol.ProjectKind {
	switch kind {
	case protocol.ProjectApp:
		return protocol.ProjectLib
	case protocol.ProjectLib:
		return protocol.ProjectPackage
	}
	return protocol.ProjectLib
}

// checkChangeKind asks a plugin that declares changeKind what making the
// generated project another kind edits, and checks it answers with a refusal
// or with edits to files that exist, and writes nothing itself: the host
// applies the edits.
func checkChangeKind(c *host.Client, ws string, p protocol.Project, all []protocol.Project, r *Report) {
	to := kindChangeTarget(p.Kind)
	newRoot, _ := projectkind.Relocate(p.Root, to)
	name := "answers what making the " + string(p.Kind) + " " + article(to) + " edits (dry run)"
	before := snapshot(ws, c.Info.Manifests)
	var res protocol.ChangeKindResult
	err := c.Call(protocol.MethodChangeKind, protocol.ChangeKindParams{Project: p, To: to, NewRoot: newRoot, Projects: all}, &res, 5*time.Minute)
	if err != nil {
		r.add(name, false, "%v", err)
		return
	}
	var missing []string
	for _, f := range res.Files {
		if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(f.Path))); err != nil {
			missing = append(missing, f.Path)
		}
	}
	switch {
	case res.Refused == "" && len(res.Files) == 0:
		r.add(name, false, "it answered with neither a refusal nor an edit: a changed kind is recorded in a file")
	case len(missing) > 0:
		r.add(name, false, "edits files that do not exist (paths are those before the move): %v", missing)
	case !maps.Equal(before, snapshot(ws, c.Info.Manifests)):
		r.add(name, false, "it wrote to the workspace: it must only answer with the new content, and let the host write it")
	default:
		r.add(name, true, "")
	}
}

// snapshot is the content of every manifest of the workspace, by path.
func snapshot(ws string, names []string) map[string]string {
	out := map[string]string{}
	for _, m := range manifests(ws, names) {
		if raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(m))); err == nil {
			out[m] = string(raw)
		}
	}
	return out
}

// manifests finds the files named like the plugin's manifests, skipping
// dot-folders and dependency folders.
func manifests(ws string, names []string) []string {
	var out []string
	_ = filepath.WalkDir(ws, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && p != ws && (strings.HasPrefix(d.Name(), ".") || d.Name() == "node_modules" || d.Name() == "__pycache__") {
			return filepath.SkipDir
		}
		if !d.IsDir() && protocol.IsManifest(names, d.Name()) {
			rel, _ := filepath.Rel(ws, p)
			if path.Dir(filepath.ToSlash(rel)) != "." {
				out = append(out, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	return out
}

func article(kind protocol.ProjectKind) string {
	if kind == protocol.ProjectApp {
		return "an app"
	}
	return "a " + string(kind)
}

// validRef reports whether a requirement or conflict names something: a
// known level, a name or a slot.
func validRef(ref protocol.PluginRef) bool {
	return !ref.IsZero() && (ref.Level == "" || ref.Level.Known())
}
