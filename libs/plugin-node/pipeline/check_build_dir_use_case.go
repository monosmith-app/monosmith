package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// comments strips // and /* */ comments, which tsconfig allows.
var comments = regexp.MustCompile(`(?s)/\*.*?\*/|(?m)^[ \t]*//[^\n]*`)

// CheckBuildDir checks that a project that builds compiles into the
// workspace's dist/<project dir>: the outDir of its tsconfig.build.json (or
// tsconfig.json) must be there, or the build writes into the source tree
// while smith caches an empty directory.
func CheckBuildDir(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if _, builds := p.Tasks["build"]; !builds {
		return protocol.Finding{}, false
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	if out := buildScriptOutDir(dir); out != "" {
		resolved := path.Clean(path.Join(p.Root, filepath.ToSlash(out)))
		return protocol.Finding{
			Check: p.Name + " compiles into " + protocol.ProjectBuildDir(p.Root), OK: protocol.InBuildDir(resolved),
			Detail: p.Root + "/package.json builds into " + resolved + ", inside the source tree",
			Remedy: "pass --outDir \"" + protocol.BuildDirFromProject(p.Root) + "\" in the build script, and point package.json's main, types and start script there",
		}, true
	}
	for _, name := range []string{"tsconfig.build.json", "tsconfig.json"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var cfg struct {
			CompilerOptions struct {
				OutDir string `json:"outDir"`
			} `json:"compilerOptions"`
		}
		if json.Unmarshal(comments.ReplaceAll(raw, nil), &cfg) != nil || cfg.CompilerOptions.OutDir == "" {
			continue
		}
		resolved := path.Clean(path.Join(p.Root, filepath.ToSlash(cfg.CompilerOptions.OutDir)))
		return protocol.Finding{
			Check: p.Name + " compiles into " + protocol.ProjectBuildDir(p.Root), OK: protocol.InBuildDir(resolved),
			Detail: p.Root + "/" + name + " compiles into " + resolved + ", inside the source tree",
			Remedy: "set compilerOptions.outDir to \"" + protocol.BuildDirFromProject(p.Root) + "\", and point package.json's main, types and start script there",
		}, true
	}
	return protocol.Finding{}, false
}

// outDirFlag finds "--outDir <dir>" (or --outDir=<dir>) in a script.
var outDirFlag = regexp.MustCompile(`--outDir[= ]+"?([^"\s&;|]+)`)

// buildScriptOutDir is the --outDir flag of the build script of the package
// in dir, which wins over the tsconfig's outDir; "" when there is none.
func buildScriptOutDir(dir string) string {
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(raw, &manifest) != nil {
		return ""
	}
	if m := outDirFlag.FindStringSubmatch(manifest.Scripts["build"]); m != nil {
		return m[1]
	}
	return ""
}
