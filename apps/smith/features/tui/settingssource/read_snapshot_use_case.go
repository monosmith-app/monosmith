package settingssource

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/features/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/features/recipecatalog"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspacerecipe"
	toml "github.com/pelletier/go-toml/v2"
)

// ReadSnapshot reads what the Settings tab shows from the workspace at root:
// the same things `smith plugin list`, `smith config list` and `smith recipe`
// read.
func ReadSnapshot(root, version string) settingsview.Snapshot {
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

	current, err := architecture.Current(root)
	if err != nil {
		snap.Err = err
		return snap
	}
	snap.Architecture = current.Name
	// The projects say how each choice fits; a workspace that cannot be opened
	// still offers the choices.
	members, _ := architecture.WorkspaceMembers(root, version)
	for _, a := range architecture.Catalog() {
		cover := a.Cover(members)
		snap.Architectures = append(snap.Architectures, settingsview.ArchitectureChoice{
			Name: a.Name, Meaning: a.Meaning, Writes: a.Writes(),
			AppliesTo: a.AppliesText(), Here: cover.Here(), Warning: a.Warning(cover),
		})
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
		snap.Settings = append(snap.Settings, viewSetting(f, value, set))
	}

	if facts, err := workspacerecipe.Load(root, version); err != nil {
		snap.WorkspaceRecipeErr = err.Error()
	} else if readings := workspacerecipe.ReadEach(facts); len(readings) == 0 {
		snap.WorkspaceRecipeErr = workspacerecipe.ErrNoProjects.Error()
	} else {
		for _, r := range readings {
			snap.WorkspaceRecipes = append(snap.WorkspaceRecipes, settingsview.WorkspaceRecipe{Language: r.Language, Projects: r.Projects, Stack: r.StackName, Fixed: r.Fixed})
		}
	}
	if listed, err := recipecatalog.ListRecipes(root); err == nil {
		for _, r := range listed {
			snap.Recipes = append(snap.Recipes, settingsview.Recipe{Path: r.Path, Name: r.Name, Version: r.Version, Extends: r.Extends, Error: r.Error})
		}
	}
	return snap
}

// viewSetting is a setting and its value as the Settings tab shows them.
func viewSetting(f configedit.Field, value string, set bool) settingsview.Setting {
	return settingsview.Setting{
		Path: f.Path, Type: f.Type, Explanation: f.Explanation(), Default: f.Default(), Reason: f.Reason, Choices: viewOptions(f.Options),
		Value: value, Set: set, Current: decode(value, set),
	}
}

// ReadProject reads what the Project section shows for one project of the
// workspace at root: what smith config list -p reads.
//
// Parameters: root is the workspace root; p is the project as the Projects
// tab knows it.
// Returns: the view; its Err says why it could not be read.
func ReadProject(root string, p settingsview.ProjectRef) settingsview.ProjectView {
	v, err := configedit.ReadProjectSettings(root, configedit.ProjectFacts{Name: p.Name, Root: p.Root, Language: p.Language, Kind: p.Kind, Version: p.Version})
	if err != nil {
		return settingsview.ProjectView{Name: p.Name, Err: err}
	}
	out := settingsview.ProjectView{Name: v.Name, Manifest: v.Manifest}
	for _, f := range v.Facts {
		out.Facts = append(out.Facts, settingsview.Fact{Key: f.Key, Value: f.Value, Explanation: f.Explanation, Source: f.Source})
	}
	for _, st := range v.Settings {
		s := viewSetting(st.Field, st.Value, st.Set)
		s.Name = st.Key()
		out.Settings = append(out.Settings, s)
	}
	return out
}

// viewOptions are a setting's options as the Settings tab shows them.
func viewOptions(options []configedit.Option) []settingsview.Option {
	var out []settingsview.Option
	for _, o := range options {
		out = append(out, settingsview.Option{Value: o.Value, Description: o.Description, Default: o.Default})
	}
	return out
}

// decode is a value as written in smith.toml as plain text: one entry for a
// scalar, one per item of an array; nothing when the setting is not set.
func decode(raw string, set bool) []string {
	if !set {
		return nil
	}
	var doc struct{ V any }
	if err := toml.Unmarshal([]byte("v = "+raw), &doc); err != nil { // words:literal
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
