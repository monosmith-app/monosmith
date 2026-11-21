package settingsview

// Snapshot is what the tab shows: the workspace's plugins, settings and
// recipes as they are on disk now.
type Snapshot struct {
	// Err says why the workspace could not be read; the rest is then empty.
	Err error
	// Plugins are the built-in language plugins; the workspace enables all of
	// them unless smith.toml lists some.
	Plugins []Plugin
	// Addons are the built-in toolchain plugins that are always loaded.
	Addons []string
	// External are the plugins smith.toml declares, with their trust.
	External []ExternalPlugin
	// Installed are the plugins found on the search path (SearchPath), the
	// unusable ones with their Problem.
	Installed  []InstalledPlugin
	SearchPath []string
	// Settings are every setting `smith config list` knows.
	Settings []Setting
	Recipes  []Recipe
	// Architectures are the choices of the Architecture section, and
	// Architecture the name of the one the workspace uses.
	Architectures []ArchitectureChoice
	Architecture  string
}

// ArchitectureChoice is a way the workspace's code can be laid out.
type ArchitectureChoice struct {
	Name string
	// Meaning is one line in plain words; Writes what choosing it changes in smith.toml.
	Meaning, Writes string
	// AppliesTo is the languages it checks ("go, node"); Here says which of
	// this workspace's projects it would check and which it would skip (empty
	// when the projects are not known); Warning is set when it would check
	// none of them, and is shown before the choice is applied.
	AppliesTo, Here, Warning string
}

// Plugin is a built-in language plugin and whether the workspace uses it.
type Plugin struct {
	Name    string
	Enabled bool
}

// ExternalPlugin is a plugin smith.toml declares.
type ExternalPlugin struct {
	Name string
	// Source says where it comes from ("smith.toml", ".smith/plugins", ...).
	Source  string
	Command string
	// Env are the variables it may see beyond the baseline.
	Env     []string
	Trusted bool
	Pinned  bool
	// Detail says why it is not trusted ("the file changed"...).
	Detail string
	// Problem is set when smith cannot even measure it.
	Problem string
}

// InstalledPlugin is a plugin on the search path.
type InstalledPlugin struct {
	Name, Version, Dir string
	// Problem is set for a directory whose plugin.json cannot be used.
	Problem string
}

// Setting is one setting of smith.toml as the schema describes it, and its
// value now.
type Setting struct {
	// Path is "table.key", e.g. "ci.devcontainer", or projects.<name>.<key>
	// for a project's setting.
	Path string
	// Name is what the row shows when it is not the path: the key of a
	// project's setting (tag_prefix). Empty shows Path.
	Name string
	// Type is the schema's: boolean, integer, string or array.
	Type string
	// Explanation is what the setting does; Default what it is when unset.
	Explanation string
	Default     string
	// Reason is one line saying why the default is what it is; empty when
	// the schema gives none.
	Reason string
	// Choices are the values of a setting with a fixed set of values (one
	// word, or the words a list may hold), each explained, the default
	// marked; empty for a setting that takes any value of its type.
	Choices []Option
	// Value is the value as written in smith.toml; Set is false when the
	// setting has its default.
	Value string
	Set   bool
	// Current is Value as plain text: one entry for a scalar, one per item of
	// an array.
	Current []string
}

// Option is one value of a setting with a fixed set of values.
type Option struct {
	// Value is the word written in smith.toml; Description says what it does.
	Value, Description string
	// Default is true for the value the setting takes when it is not set.
	Default bool
}

// values are the words of the options, in their order.
func (st Setting) values() []string {
	out := make([]string, 0, len(st.Choices))
	for _, o := range st.Choices {
		out = append(out, o.Value)
	}
	return out
}

// label is what the row of a setting shows: its name, else its path.
func (st Setting) label() string {
	if st.Name != "" {
		return st.Name
	}
	return st.Path
}

// ProjectRef is the project whose settings the Project section shows, as the
// Projects tab knows it.
type ProjectRef struct {
	// Name is the project's name; Root its workspace-relative folder.
	Name, Root string
	// Language, Kind and Version are what its plugin reports.
	Language, Kind, Version string
}

// ProjectView is what the Project section shows: the values the project's
// manifest owns, read-only, and the settings of [projects.<name>].
type ProjectView struct {
	// Err says why the project could not be read; the rest is then empty.
	Err error
	// Name is the project's name; Manifest the file that owns the facts.
	Name, Manifest string
	// Facts are the read-only values, with the file that holds each.
	Facts []Fact
	// Settings are the settings of [projects.<name>] that apply to it.
	Settings []Setting
}

// Fact is a value the project's manifest or plugin owns.
type Fact struct {
	// Key names it (name, version, language, kind, root); Value is the value.
	Key, Value string
	// Explanation says what it is and where to change it; Source is the file.
	Explanation, Source string
}

// Recipe is a recipe file of the workspace.
type Recipe struct {
	Path, Name, Version, Extends string
	// Error says why the file is not a valid recipe.
	Error string
}
