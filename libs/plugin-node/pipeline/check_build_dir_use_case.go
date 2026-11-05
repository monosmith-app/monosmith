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
	finding := protocol.Finding{
		Check: p.Name + " compiles into " + protocol.ProjectBuildDir(p.Root), OK: protocol.InBuildDir(resolved),
		Detail: source + " compiles into " + resolved + ", inside the source tree",
		Remedy: remedy,
	}
	if !finding.OK {
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
