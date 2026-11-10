package pipeline

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckBuildDir checks that a project that builds compiles into the
// workspace's dist/<project dir>: the --outDir of its build script, or else
// the outDir of its tsconfig.build.json (or tsconfig.json), must be there, or
// the build writes into the source tree while smith caches an empty
// directory.
func CheckBuildDir(workspaceRoot string, p protocol.Project, dependents ...string) (protocol.Finding, bool) {
	if _, builds := p.Tasks["build"]; !builds {
		return protocol.Finding{}, false
	}
	if v, _ := p.Settings["vite"]["output"].(string); v == "project" {
		// A web app built per environment that keeps its builds in its own
		// dist/<env> (ADR 0081): the owner chose it when the app was added.
		return protocol.Finding{}, false
	}
	config, outDir, ok := manifestfixes.LocateOutDir(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
	if !ok {
		return protocol.Finding{}, false
	}
	resolved := path.Clean(path.Join(p.Root, filepath.ToSlash(outDir)))
	source := p.Root + "/" + config
	remedy := "set compilerOptions.outDir to \"" + protocol.BuildDirFromProject(p.Root) + "\", and point package.json's main, types and start script there"
	if config == "" {
		source = p.Root + "/package.json"
		remedy = "pass --outDir \"" + protocol.BuildDirFromProject(p.Root) + "\" in the build script, and point package.json's main, types and start script there"
	}
	if p.Kind == protocol.ProjectPackage {
		remedy = libraryRemedy(p)
	}
	finding := protocol.Finding{
		Check: p.Name + " compiles into " + protocol.ProjectBuildDir(p.Root), Project: p.Name, OK: protocol.InBuildDir(resolved),
		Detail: source + " compiles into " + resolved + ", inside the source tree",
		Remedy: remedy,
	}
	if !finding.OK {
		finding.Problem = "compiles into " + resolved + ", not " + protocol.ProjectBuildDir(p.Root)
		broken := false
		if declared := declaredBuildOutputs(p); declared != "" && !strings.Contains(", "+declared+",", ", "+resolved+",") {
			finding.Detail += ". The build task caches " + declared + " instead, which that build never writes, so a cache hit restores nothing"
			broken = true
		}
		fix, refusal := manifestfixes.MigrateBuildFix(workspaceRoot, p, config, outDir, dependents)
		finding.Fix = fix
		if refusal != "" {
			finding.Remedy = "smith does not move this build for you, because " + refusal + ". To move it by hand: " + remedy
			// A refused move leaves a build that works, so it is a warning;
			// only a cache that restores nothing is a failure.
			finding.Warning = !broken
		}
	}
	return finding, true
}

// libraryRemedy says how a library package builds into the workspace's dist
// by hand. Its package.json keeps main, types and exports in the published
// form (./dist/...): a path outside the package, such as ../../dist/..., is
// not allowed in exports and cannot be packed.
func libraryRemedy(p protocol.Project) string {
	target := protocol.BuildDirFromProject(p.Root)
	return "build into \"" + target + "\": the --outfile, --outDir, --declarationDir and --tsBuildInfoFile of the build script and the outDir, declarationDir and tsBuildInfoFile of its tsconfig files. " +
		"Keep main, types and exports in package.json as they are (./dist/...): smith packs from " + protocol.ProjectBuildDir(p.Root) + " and makes them relative to it. " +
		"Map the package name to its source in tsconfig.base.json \"paths\" (\"name\": [\"" + p.Root + "/src/index.ts\"]) and in Jest, so importers in the repository read the source, not a build"
}

// CheckLibraryBuildDir is CheckBuildDir with the workspace's plan for library
// packages: a package the plan moves carries its one grouped fix, and one it
// refuses says why. A project the plan does not cover is judged as
// CheckBuildDir does.
func CheckLibraryBuildDir(workspaceRoot string, p protocol.Project, plan manifestfixes.LibraryPlan, dependents ...string) (protocol.Finding, bool) {
	f, ok := CheckBuildDir(workspaceRoot, p, dependents...)
	if !ok || f.OK {
		return f, ok
	}
	if plan.Moves(p.Name) {
		f.Fix = plan.Fix
		f.Remedy = "smith doctor --fix --only \"" + manifestfixes.LibraryBuildDescription + "\" moves the build of " + p.Name + " and the other libraries into dist/<project dir> and shows every changed file first"
	} else if reason, refused := plan.Refusals[p.Name]; refused {
		f.Fix = nil
		f.Remedy = "smith does not move this build for you, because " + reason + ". To move it by hand: " + libraryRemedy(p)
	}
	return f, ok
}

// dependentsOf names the projects that import p by one of its package names.
func dependentsOf(projects []protocol.Project, p protocol.Project) []string {
	var out []string
	for _, q := range projects {
		if q.Name == p.Name {
			continue
		}
		for _, name := range p.Exports {
			if slices.Contains(q.Imports, name) {
				out = append(out, q.Name)
				break
			}
		}
	}
	return out
}

// declaredBuildOutputs names the folders the build task caches, from the
// workspace root, or "" when it declares none.
func declaredBuildOutputs(p protocol.Project) string {
	t, ok := p.Tasks["build"]
	if !ok || len(t.Outputs) == 0 {
		return ""
	}
	resolved := make([]string, len(t.Outputs))
	for i, o := range t.Outputs {
		resolved[i] = path.Clean(path.Join(p.Root, o))
	}
	return strings.Join(resolved, ", ")
}
