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
	// Path is "table.key", e.g. "ci.devcontainer".
	Path string
	// Type is the schema's: boolean, integer, string or array.
	Type string
	// Explanation is what the setting does; Default what it is when unset.
	Explanation string
	Default     string
	// Choices are the values an array setting may hold.
	Choices []string
	// Value is the value as written in smith.toml; Set is false when the
	// setting has its default.
	Value string
	Set   bool
	// Current is Value as plain text: one entry for a scalar, one per item of
	// an array.
	Current []string
}

// Recipe is a recipe file of the workspace.
type Recipe struct {
	Path, Name, Version, Extends string
	// Error says why the file is not a valid recipe.
	Error string
}
