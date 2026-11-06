package turboplan

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

// managerFiles are the files that belong to a package manager other than npm
// and that smith install does not use.
var managerFiles = map[string]string{
	"pnpm-lock.yaml": "pnpm's lockfile", "yarn.lock": "yarn's lockfile", "bun.lock": "bun's lockfile",
	"bun.lockb": "bun's lockfile", ".yarnrc.yml": "yarn's configuration", ".yarnrc": "yarn's configuration",
	".pnp.cjs": "yarn Plug'n'Play's resolver", ".pnp.loader.mjs": "yarn Plug'n'Play's loader",
}

// lockfiles are managerFiles that hold resolved versions.
var lockfiles = map[string]bool{"pnpm-lock.yaml": true, "yarn.lock": true, "bun.lock": true, "bun.lockb": true}

// pnpmOnlyNpmrc are .npmrc keys only pnpm reads.
var pnpmOnlyNpmrc = []string{
	"shamefully-hoist", "public-hoist-pattern", "hoist-pattern", "hoist", "node-linker", "auto-install-peers",
	"link-workspace-packages", "prefer-workspace-packages", "strict-peer-dependencies", "side-effects-cache",
	"enable-pre-post-scripts", "use-node-version", "shared-workspace-lockfile", "store-dir", "package-import-method",
	"inject-workspace-packages", "dedupe-peer-dependents", "resolution-mode", "recursive-install",
}

// pnpmFieldNote says what to do with the settings pnpm keeps in package.json
// or pnpm-workspace.yaml besides overrides.
const pnpmFieldNote = " has no npm equivalent and is dropped with the pnpm field"

var registrySettings = regexp.MustCompile(`(?m)^\s*(npmRegistryServer|npmScopes|npmAuthToken|npmAlwaysAuth|npmPublishRegistry)\b`)

// planPackageManager turns the workspace's pnpm or yarn setup into the npm
// workspaces smith install uses, and reports every file and setting it
// removes or cannot carry.
func planPackageManager(plan *convertplan.Plan, w *turboworkspace.Workspace, files []string, read func(string) string) {
	tracked := map[string]bool{}
	for _, f := range files {
		tracked[f] = true
	}
	plan.Add(convertplan.Mapped, "", "install", "dependencies are installed with smith install (smith i), which runs npm install for the workspace's Node projects, so "+managerLabel(w.Manager)+" is not needed")

	planWorkspaceGlobs(plan, w, tracked)
	if w.RootExtras.PackageManager != "" {
		plan.RemoveFields = append(plan.RemoveFields, "packageManager")
		plan.Add(convertplan.Dropped, "", "packageManager", w.RootExtras.PackageManager+": smith installs with npm, so the field no longer pins anything")
	}
	for _, name := range []string{"pnpm", "yarn", "bun"} {
		if _, has := w.RootExtras.Engines[name]; has {
			plan.RemoveFields = append(plan.RemoveFields, "engines."+name)
			plan.Add(convertplan.Dropped, "", "engines."+name, "the "+name+" version requirement; npm does not check it")
		}
	}
	planOverrides(plan, w)

	lock := false
	for _, f := range sortedFiles(managerFiles) {
		if !tracked[f] {
			continue
		}
		plan.DeletePath(f)
		lock = lock || lockfiles[f]
		plan.Add(convertplan.Dropped, "", f, managerFiles[f]+"; smith install uses npm")
	}
	if lock {
		plan.Add(convertplan.Changed, "", "package-lock.json", "the lockfile's resolved versions are not carried over: the conversion's npm install resolves every range again and writes package-lock.json, so a dependency may move to a newer version inside its range; review the diff before committing")
	}
	if hasPrefix(files, ".yarn/") {
		plan.DeletePath(".yarn")
		plan.Add(convertplan.Dropped, "", ".yarn", "yarn's release, plugins and cache")
		if hasPrefix(files, ".yarn/patches/") {
			plan.Add(convertplan.Unsupported, "", ".yarn/patches", "yarn patches have no npm equivalent; apply them with patch-package")
		}
	}
	if text := read(".yarnrc.yml"); text != "" {
		if registrySettings.MatchString(text) {
			plan.Add(convertplan.Manual, "", ".yarnrc.yml", "sets a registry, scope or credentials; npm reads them from .npmrc, so copy them there")
		}
		if strings.Contains(text, "nodeLinker: pnp") {
			plan.Add(convertplan.Changed, "", ".yarnrc.yml", "nodeLinker: pnp: dependencies are installed into node_modules, as npm does")
		}
	}
	if tracked[".pnpmfile.cjs"] {
		plan.Add(convertplan.Unsupported, "", ".pnpmfile.cjs", "pnpm install hooks are not run by npm; the file is kept for reference")
	}
	if text := read(".npmrc"); text != "" {
		var found []string
		for _, key := range pnpmOnlyNpmrc {
			if regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key) + `\s*=`).MatchString(text) {
				found = append(found, key)
			}
		}
		if len(found) > 0 {
			plan.Add(convertplan.Changed, "", ".npmrc", "kept, but npm ignores the pnpm settings "+strings.Join(found, ", ")+"; remove them")
		}
	}
	if w.Pnpm != nil {
		for _, key := range w.Pnpm.Keys {
			switch key {
			case "packages", "catalog", "catalogs", "overrides":
			default:
				plan.Add(convertplan.Unsupported, "", "pnpm-workspace.yaml", key+pnpmFieldNote)
			}
		}
	}
	planSpecifiers(plan, w)
}

func managerLabel(m turboworkspace.Manager) string {
	if m == turboworkspace.NPM {
		return "a second package manager"
	}
	return string(m)
}

// planWorkspaceGlobs makes the root package.json "workspaces" the one list of
// packages, as npm reads it.
func planWorkspaceGlobs(plan *convertplan.Plan, w *turboworkspace.Workspace, tracked map[string]bool) {
	if len(w.Packages) == 0 {
		plan.Add(convertplan.Blocker, "", "workspaces", "no package matches the workspace globs "+list(w.Globs)+" (pnpm-workspace.yaml or package.json workspaces): there is nothing to convert")
	}
	if w.Pnpm != nil && len(w.Pnpm.Packages) > 0 {
		plan.SetWorkspaces = w.Pnpm.Packages
		plan.AddEdit("", "pnpm-workspace.yaml", "its packages become the package.json \"workspaces\" field: "+strings.Join(w.Pnpm.Packages, ", "))
	}
	if tracked["pnpm-workspace.yaml"] {
		plan.DeletePath("pnpm-workspace.yaml")
		plan.Add(convertplan.Dropped, "", "pnpm-workspace.yaml", "removed once its packages are in package.json")
	}
	var object struct {
		Packages []string `json:"packages"`
		Nohoist  []string `json:"nohoist"`
	}
	if json.Unmarshal(w.RootPackage.Workspaces, &object) == nil && len(w.RootPackage.Workspaces) > 0 {
		plan.SetWorkspaces = object.Packages
		plan.Add(convertplan.Changed, "", "workspaces", "the object form (yarn's {packages, nohoist}) becomes the list of globs npm reads")
		if len(object.Nohoist) > 0 {
			plan.Add(convertplan.Unsupported, "", "workspaces", "nohoist "+list(object.Nohoist)+" has no npm equivalent and is dropped")
		}
	}
}

// lockfilePlatformPackages are the wasm fallbacks a lockfile made on one OS
// leaves out; smith doctor asks for them as direct dependencies.
var lockfilePlatformPackages = map[string]bool{"@emnapi/core": true, "@emnapi/runtime": true}

// planOverrides moves pnpm's and yarn's dependency overrides to npm's.
func planOverrides(plan *convertplan.Plan, w *turboworkspace.Workspace) {
	type source struct {
		where string
		from  map[string]json.RawMessage
	}
	sources := []source{{"resolutions", w.RootExtras.Resolutions}}
	if raw, ok := w.RootExtras.Pnpm["overrides"]; ok {
		var m map[string]json.RawMessage
		_ = json.Unmarshal(raw, &m)
		sources = append(sources, source{"pnpm.overrides", m})
	}
	if w.Pnpm != nil && len(w.Pnpm.Overrides) > 0 {
		m := map[string]json.RawMessage{}
		for k, v := range w.Pnpm.Overrides {
			m[k], _ = json.Marshal(v)
		}
		sources = append(sources, source{"pnpm-workspace.yaml overrides", m})
	}
	for _, s := range sources {
		keys := make([]string, 0, len(s.from))
		for k := range s.from {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			var value string
			_ = json.Unmarshal(s.from[k], &value)
			switch {
			case value == "" || !plainOverrideKey(k) || strings.HasPrefix(value, "link:") || strings.HasPrefix(value, "workspace:") || strings.HasPrefix(value, "catalog:") || strings.HasPrefix(value, "patch:"):
				plan.Add(convertplan.Manual, "", s.where, fmt.Sprintf("%q: %s uses a selector or specifier npm overrides do not have; write the override by hand", k, string(s.from[k])))
			default:
				if plan.SetOverrides == nil {
					plan.SetOverrides = map[string]string{}
				}
				if lockfilePlatformPackages[k] {
					// npm refuses an override that differs from a direct dependency, and
					// a lockfile made on one OS needs these as direct ones: a devDependency
					// plus an override that follows it.
					if plan.AddDevDependencies == nil {
						plan.AddDevDependencies = map[string]string{}
					}
					plan.AddDevDependencies[k] = value
					plan.SetOverrides[k] = "$" + k
					plan.AddEdit("", s.where, fmt.Sprintf("%s: %s becomes a devDependency and the npm override \"$%s\" (npm refuses an override that differs from a direct dependency)", k, value, k))
					continue
				}
				plan.SetOverrides[k] = value
				plan.AddEdit("", s.where, fmt.Sprintf("%s: %s becomes an npm override", k, value))
			}
		}
	}
	if len(w.RootExtras.Resolutions) > 0 {
		plan.RemoveFields = append(plan.RemoveFields, "resolutions")
	}
	if len(w.RootExtras.Pnpm) > 0 {
		plan.RemoveFields = append(plan.RemoveFields, "pnpm")
		keys := make([]string, 0, len(w.RootExtras.Pnpm))
		for k := range w.RootExtras.Pnpm {
			if k != "overrides" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			plan.Add(convertplan.Unsupported, "", "package.json pnpm."+k, "pnpm."+k+pnpmFieldNote)
		}
	}
}

// planSpecifiers rewrites the dependency ranges npm does not accept.
func planSpecifiers(plan *convertplan.Plan, w *turboworkspace.Workspace) {
	type manifest struct {
		dir, label string
		json       nxworkspace.PackageJSON
	}
	manifests := []manifest{{".", "", w.RootPackage}}
	for _, p := range w.Packages {
		manifests = append(manifests, manifest{p.Dir, p.Dir, p.JSON})
	}
	var examples []string
	count := 0
	for _, m := range manifests {
		project := ""
		if pp := plan.ProjectAt(m.dir); pp != nil {
			project = pp.Name
		}
		for _, deps := range []map[string]string{m.json.Dependencies, m.json.DevDependencies, m.json.PeerDependencies, m.json.OptionalDependencies} {
			names := make([]string, 0, len(deps))
			for n := range deps {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				s, ok := SpecifierOf(w, n, deps[n])
				switch {
				case !ok:
				case s.Problem != "":
					plan.Add(convertplan.Manual, project, n, s.Problem)
				default:
					plan.SetSpecifier(m.dir, n, s.Range)
					count++
					if len(examples) < 3 {
						examples = append(examples, n+" "+deps[n]+" to "+s.Range)
					}
					if s.Approximate {
						plan.Add(convertplan.Changed, project, n, deps[n]+" becomes * because the package declares no version to build a range from")
					}
				}
			}
		}
	}
	if count > 0 {
		plan.AddEdit("", "dependency specifiers", fmt.Sprintf("%d workspace: and catalog: ranges become what npm accepts (workspace:* is *, workspace:^ is ^<the package's version>, catalog: is its range from pnpm-workspace.yaml), e.g. %s", count, strings.Join(examples, "; ")))
	}
}

func sortedFiles(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func hasPrefix(files []string, prefix string) bool {
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

// plainOverrideKey reports whether an override key is a bare package name,
// which npm overrides take; pnpm's "a>b" and "foo@<2" selectors and yarn's
// "**/foo" and "a/b" paths are not.
func plainOverrideKey(k string) bool {
	if strings.ContainsAny(k, "> *") {
		return false
	}
	if strings.HasPrefix(k, "@") {
		scope, name, ok := strings.Cut(k, "/")
		return ok && !strings.Contains(name, "/") && !strings.Contains(name, "@") && scope != ""
	}
	return !strings.ContainsAny(k, "/@")
}
