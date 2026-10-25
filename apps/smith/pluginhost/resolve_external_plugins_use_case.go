package pluginhost

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ResolveExternal works out what smith would run for an external plugin the
// workspace declares: the command, where it comes from, and the environment
// variables it asks for (its manifest's, and smith.toml's grant).
func ResolveExternal(workspaceRoot string, external map[string]workspace.ExternalPlugin, name string) (pluginadmit.Plugin, error) {
	cfg := external[name]
	p := pluginadmit.Plugin{Name: name, WorkspaceEnv: cfg.Env, Pin: cfg.SHA256}
	if len(cfg.Command) > 0 {
		p.Command, p.Source, p.Where = cfg.Command, workspace.FileName, filepath.Join(workspaceRoot, workspace.FileName)
		return p, nil
	}
	found, err := FindPlugin(workspaceRoot, name)
	if err != nil {
		return pluginadmit.Plugin{}, err
	}
	p.Command, p.ManifestEnv = found.Argv, found.Env
	p.Source, p.Where = sourceOf(workspaceRoot, found.Dir), found.Dir
	return p, nil
}

// ResolveAllExternal resolves every external plugin the workspace declares in
// smith.toml, by name: those are the ones it needs. A plugin installed on the
// search path but not declared is never started.
func ResolveAllExternal(workspaceRoot string, external map[string]workspace.ExternalPlugin) ([]pluginadmit.Plugin, error) {
	names := make([]string, 0, len(external))
	for name := range external {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]pluginadmit.Plugin, 0, len(names))
	for _, name := range names {
		p, err := ResolveExternal(workspaceRoot, external, name)
		if err != nil {
			return nil, fmt.Errorf("smith.toml [plugins.external.%s]: %w", name, err)
		}
		out = append(out, p)
	}
	return out, nil
}

// sourceOf names the search-path entry an installed plugin's directory is
// under.
func sourceOf(workspaceRoot, dir string) string {
	parent := filepath.Dir(dir)
	if workspaceRoot != "" && sameDir(parent, filepath.Join(workspaceRoot, ".smith", "plugins")) {
		return ".smith/plugins"
	}
	for _, d := range filepath.SplitList(os.Getenv("SMITH_PLUGIN_PATH")) {
		if d != "" && sameDir(parent, d) {
			return "SMITH_PLUGIN_PATH"
		}
	}
	return "$SMITH_HOME/plugins"
}

func sameDir(a, b string) bool {
	if x, err := filepath.Abs(a); err == nil {
		a = x
	}
	if x, err := filepath.Abs(b); err == nil {
		b = x
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
