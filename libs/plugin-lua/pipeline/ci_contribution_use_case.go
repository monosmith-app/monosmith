package pipeline

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-lua/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// luaTools installs Lua, LuaRocks and the two tools the tasks run, on every
// provider, from the distribution's packages (a Debian or Ubuntu runner).
var luaTools = protocol.Toolchain{
	Name: "lua", Version: "5.4",
	Install: "SUDO=$(command -v sudo || true)\n" +
		"command -v luarocks >/dev/null || { $SUDO apt-get update -qq && $SUDO apt-get install -y -qq lua5.4 luarocks; }\n" +
		"command -v luacheck >/dev/null || $SUDO luarocks install luacheck\n" +
		"command -v busted >/dev/null || $SUDO luarocks install busted",
}

// CI returns the Lua contribution: Lua with LuaRocks, luacheck and busted.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	return protocol.CIResult{Toolchains: []protocol.Toolchain{luaTools}}
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that LuaRocks and luacheck are installed, and warns when a
// rockspec's file name does not say its package and version, which LuaRocks
// expects.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, rocksErr := probe(workspaceRoot, []string{"luarocks", "--version"})
	_, checkErr := probe(workspaceRoot, []string{"luacheck", "--version"})
	findings := []protocol.Finding{
		{Check: "LuaRocks is installed", OK: rocksErr == nil, Detail: "luarocks is not on PATH", Remedy: "install LuaRocks (https://luarocks.org/#quick-start)"},
		{Check: "luacheck is installed", OK: checkErr == nil, Detail: "luacheck is not on PATH", Remedy: "run luarocks install luacheck"},
	}
	for _, p := range params.Projects {
		if f, ok := rockspecNameFinding(workspaceRoot, p); ok {
			findings = append(findings, f)
		}
	}
	return protocol.DoctorResult{Findings: findings}
}

// rockspecNameFinding compares a project's rockspec file name with the
// package-version.rockspec LuaRocks expects. The fix renames the file.
func rockspecNameFinding(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "*.rockspec"))
	if len(found) == 0 {
		return protocol.Finding{}, false
	}
	src, err := os.ReadFile(found[0])
	if err != nil {
		return protocol.Finding{}, false
	}
	spec := discovery.ParseRockspec(string(src))
	if spec.Package == "" || spec.Version == "" {
		return protocol.Finding{}, false
	}
	have, want := filepath.Base(found[0]), spec.Package+"-"+spec.Version+".rockspec"
	finding := protocol.Finding{Check: p.Name + "'s rockspec is named for its version", OK: have == want, Warning: true}
	if !finding.OK {
		old, renamed := path.Join(p.Root, have), path.Join(p.Root, want)
		finding.Detail = old + " should be " + want + ", as LuaRocks expects"
		finding.Remedy = "rename " + old + " to " + renamed
		finding.Fix = &protocol.Fix{
			Description: "rename the rockspec to " + want,
			AppliesTo:   []string{p.Name},
			Files:       []protocol.FileEdit{{Path: old, Delete: true}, {Path: renamed, Content: string(src)}},
		}
	}
	return finding, true
}
