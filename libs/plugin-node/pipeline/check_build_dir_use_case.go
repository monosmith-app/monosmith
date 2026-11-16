package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckBuildDir checks where a project that builds compiles (ADR 0102). An app
// compiles into the workspace's dist/<project dir>, its deploy output; a
// library or package compiles inside its own folder (dist), so package.json's
// main, types and exports lead to files of the package for every tool. The
// folder is the --outDir of the build script, or else the outDir of its
// tsconfig.build.json (or tsconfig.json).
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
	if config == "" {
		source = p.Root + "/package.json"
	}
	if !discovery.BuildsLikeApp(p.Kind, p.Stack == discovery.VSCodeExtensionStack || isExtension(workspaceRoot, p)) {
		return checkLibraryBuildDir(workspaceRoot, p, config, outDir, resolved, source)
	}
	remedy := "set compilerOptions.outDir to \"" + protocol.BuildDirFromProject(p.Root) + "\", and point package.json's main, types and start script there"
	if config == "" {
		remedy = "pass --outDir \"" + protocol.BuildDirFromProject(p.Root) + "\" in the build script, and point package.json's main, types and start script there"
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
		}
		// A build that works where it is is a warning; a cache that restores
		// nothing, or an app smith can move, is a failure.
		finding.Warning = !broken && refusal != ""
	}
	return finding, true
}

// checkLibraryBuildDir is the check of a library or package: its build stays
// inside its folder. A build outside it fails, because main, types and exports
// would have to leave the package (../../dist/...): tools read them through the
// link node_modules/<name>, where ../.. leads elsewhere, and a scoped name adds
// one more level.
func checkLibraryBuildDir(workspaceRoot string, p protocol.Project, config, outDir, resolved, source string) (protocol.Finding, bool) {
	if p.Root == "." || p.Root == "" {
		return protocol.Finding{}, false
	}
	finding := protocol.Finding{
		Check: p.Name + " builds inside its own folder", Project: p.Name, OK: manifestfixes.InsideProject(p.Root, resolved),
		Detail: source + " compiles into " + resolved + ", outside the package. package.json main, types and exports cannot lead there: TypeScript, Jest and Vitest read them through node_modules/" + cmpName(p) + ", where ../.. reaches another folder",
		Remedy: "build into dist inside " + p.Root + " (the outDir of its tsconfig, or the --outDir of its build script) and point main, types and exports at ./dist/...",
	}
	if !finding.OK {
		finding.Problem = "compiles into " + resolved + ", outside its folder"
		finding.Fix = manifestfixes.BuildInPlaceFix(workspaceRoot, p, config, outDir)
	}
	return finding, true
}

// isExtension reports whether a project's package.json names a VS Code engine.
func isExtension(workspaceRoot string, p protocol.Project) bool {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Engines struct {
			VSCode string `json:"vscode"`
		} `json:"engines"`
	}
	return json.Unmarshal(raw, &pkg) == nil && pkg.Engines.VSCode != ""
}

// cmpName is the package name of a project, or its name when it exports none.
func cmpName(p protocol.Project) string {
	if p.PackageName != "" {
		return p.PackageName
	}
	return p.Name
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
