package pluginhost

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// SearchPath lists the directories installed plugins are looked for in, in
// order, the first match winning: the workspace's .smith/plugins, each
// directory of SMITH_PLUGIN_PATH, then $SMITH_HOME/plugins (SMITH_HOME defaults
// to ~/.smith). None of it depends on where or how smith itself is installed.
func SearchPath(workspaceRoot string) []string {
	var dirs []string
	if workspaceRoot != "" {
		dirs = append(dirs, filepath.Join(workspaceRoot, ".smith", "plugins"))
	}
	for _, d := range filepath.SplitList(os.Getenv("SMITH_PLUGIN_PATH")) {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	if home := Home(); home != "" {
		dirs = append(dirs, filepath.Join(home, "plugins"))
	}
	return dirs
}

// Home is smith's per-user directory: SMITH_HOME, or ~/.smith.
func Home() string { return plugintrust.Home() }

// InstalledProblem is a plugin directory whose plugin.json cannot be used.
type InstalledProblem struct {
	// Dir is the plugin's directory.
	Dir string
	// Problem says what is wrong with its plugin.json.
	Problem string
}

// FindInstalled lists the plugins on the search path, one per name: a plugin
// earlier on the path hides one of the same name later. Each is a directory
// holding a plugin.json, directly under a search directory. A directory whose
// plugin.json is invalid is left out (FindInstalledReport says which).
func FindInstalled(workspaceRoot string) ([]Installed, error) {
	out, _, err := FindInstalledReport(workspaceRoot)
	return out, err
}

// FindInstalledReport is FindInstalled, and also the directories it skipped
// because their plugin.json is invalid: one broken plugin does not hide the
// others.
func FindInstalledReport(workspaceRoot string) ([]Installed, []InstalledProblem, error) {
	seen := map[string]bool{}
	var out []Installed
	var problems []InstalledProblem
	for _, dir := range SearchPath(workspaceRoot) {
		entries, err := os.ReadDir(dir)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			p, err := readInstalled(filepath.Join(dir, e.Name()))
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				problems = append(problems, InstalledProblem{Dir: filepath.Join(dir, e.Name()), Problem: err.Error()})
				continue
			}
			if !seen[p.Name] {
				seen[p.Name] = true
				out = append(out, p)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, problems, nil
}

// FindPlugin returns the installed plugin with a name.
func FindPlugin(workspaceRoot, name string) (Installed, error) {
	all, problems, err := FindInstalledReport(workspaceRoot)
	if err != nil {
		return Installed{}, err
	}
	for _, p := range all {
		if p.Name == name {
			return p, nil
		}
	}
	skipped := ""
	for _, p := range problems {
		skipped += "; skipped: " + p.Problem
	}
	return Installed{}, fmt.Errorf("plugin %q is not installed: no %s/%s in %s; give its command in smith.toml, or install it there%s", name, name, ManifestFile, strings.Join(SearchPath(workspaceRoot), ", "), skipped)
}

// LaunchCommand is how a plugin is started by name: the command smith.toml
// gives, else the built-in one, else the one installed on the search path.
func LaunchCommand(workspaceRoot string, external map[string]workspace.ExternalPlugin, name string) ([]string, error) {
	if e, ok := external[name]; ok && len(e.Command) > 0 {
		return e.Command, nil
	}
	if _, ok := external[name]; !ok {
		if _, builtin := Builtin(name); builtin {
			self, err := os.Executable()
			if err != nil {
				return nil, err
			}
			return []string{self, "plugin", "serve", name}, nil
		}
	}
	found, err := FindPlugin(workspaceRoot, name)
	if err != nil {
		return nil, err
	}
	return found.Argv, nil
}

func readInstalled(dir string) (Installed, error) {
	raw, err := os.ReadFile(filepath.Join(dir, ManifestFile))
	if err != nil {
		return Installed{}, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Installed{}, fmt.Errorf("%s: %w", filepath.Join(dir, ManifestFile), err)
	}
	if m.Name == "" || len(m.Command) == 0 {
		return Installed{}, fmt.Errorf("%s: name and command are required", filepath.Join(dir, ManifestFile))
	}
	return Installed{Manifest: m, Dir: dir, Argv: resolveCommand(dir, m.Command)}, nil
}

func resolveCommand(dir string, command []string) []string {
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	fill := strings.NewReplacer("{os}", runtime.GOOS, "{arch}", runtime.GOARCH, "{exe}", exe)
	argv := make([]string, len(command))
	for i, a := range command {
		argv[i] = fill.Replace(a)
	}
	if strings.HasPrefix(argv[0], "./") || strings.HasPrefix(argv[0], "../") {
		argv[0] = filepath.Join(dir, filepath.FromSlash(argv[0]))
	}
	return argv
}
