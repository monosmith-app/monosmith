package pipeline

import (
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckBuildDir checks that a project that builds compiles into the
// workspace's dist/<project dir>: the --outDir of its build script, or else
// the outDir of its tsconfig.build.json (or tsconfig.json), must be there, or
// the build writes into the source tree while smith caches an empty
// directory.
func CheckBuildDir(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if _, builds := p.Tasks["build"]; !builds {
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
		finding.Fix = manifestfixes.BuildIntoDistFix(workspaceRoot, p, config, outDir)
	}
	return finding, true
}
