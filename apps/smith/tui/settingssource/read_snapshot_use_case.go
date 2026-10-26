package settingssource

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	toml "github.com/pelletier/go-toml/v2"
)

// ReadSnapshot reads what the Settings tab shows from the workspace at root:
// the same things `smith plugin list`, `smith config list` and `smith recipe`
// read.
func ReadSnapshot(root string) settingsview.Snapshot {
	cfg, err := workspace.Load(root)
	if err != nil {
		return settingsview.Snapshot{Err: err}
	}
	snap := settingsview.Snapshot{SearchPath: pluginhost.SearchPath(root)}

	// An empty list means every built-in language plugin.
	on := map[string]bool{}
	for _, n := range cfg.Plugins.Enabled {
		on[n] = true
	}
	for _, n := range pluginhost.BuiltinNames() {
		snap.Plugins = append(snap.Plugins, settingsview.Plugin{Name: n, Enabled: len(on) == 0 || on[n]})
	}
	for n := range pluginhost.BuiltinAddons {
		snap.Addons = append(snap.Addons, n)
	}
	sort.Strings(snap.Addons)

	used := map[string]bool{}
	for _, st := range pluginhost.ExternalStatuses(root, cfg) {
		used[st.Name] = true
		snap.External = append(snap.External, settingsview.ExternalPlugin{
			Name: st.Name, Source: st.Source, Command: strings.Join(st.Command, " "), Env: st.Env,
			Trusted: st.Trusted, Pinned: st.Pinned, Detail: st.Detail, Problem: st.Problem,
		})
	}
	installed, broken, err := pluginhost.FindInstalledReport(root)
	if err == nil {
		for _, p := range installed {
			if !used[p.Name] {
				snap.Installed = append(snap.Installed, settingsview.InstalledPlugin{Name: p.Name, Version: p.Version, Dir: p.Dir})
			}
		}
		for _, b := range broken {
			snap.Installed = append(snap.Installed, settingsview.InstalledPlugin{Name: filepath.Base(b.Dir), Dir: b.Dir, Problem: b.Problem})
		}
	}

	fields, err := configedit.Fields()
	if err != nil {
		snap.Err = err
		return snap
	}
	// A default stack exists per language; list the ones the file sets, as
	// smith config list does.
	var languages []string
	for language := range cfg.DefaultStacks {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	for _, language := range languages {
		fields = append(fields, configedit.DefaultStackField(language))
	}
	for _, f := range fields {
		if f.Path == "plugins.enabled" {
			// The Plugins section edits it, and keeps one plugin on.
			continue
		}
		value, set, err := configedit.Get(root, f.Path)
		if err != nil {
			snap.Err = err
			return snap
		}
		snap.Settings = append(snap.Settings, settingsview.Setting{
			Path: f.Path, Type: f.Type, Explanation: f.Explanation(), Default: f.Default(), Choices: f.Enum,
			Value: value, Set: set, Current: decode(value, set),
		})
	}

	if listed, err := recipecatalog.ListRecipes(root); err == nil {
		for _, r := range listed {
			snap.Recipes = append(snap.Recipes, settingsview.Recipe{Path: r.Path, Name: r.Name, Version: r.Version, Extends: r.Extends, Error: r.Error})
		}
	}
	return snap
}

// decode is a value as written in smith.toml as plain text: one entry for a
// scalar, one per item of an array; nothing when the setting is not set.
func decode(raw string, set bool) []string {
	if !set {
		return nil
	}
	var doc struct{ V any }
	if err := toml.Unmarshal([]byte("v = "+raw), &doc); err != nil {
		return []string{raw}
	}
	if list, ok := doc.V.([]any); ok {
		out := make([]string, 0, len(list))
		for _, v := range list {
			out = append(out, fmt.Sprint(v))
		}
		return out
	}
	return []string{fmt.Sprint(doc.V)}
}
