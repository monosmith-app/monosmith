package choicedata

// File is one choice file as it is written: choices/<name>.toml. The base
// file (base.toml) has the same members, but no axis, name or selection
// members: it is what every choice of the plugin builds on.
type File struct {
	// Axis, Name, Description and Language describe the choice (protocol.Choice).
	Axis        string `toml:"axis"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Language    string `toml:"language"`
	// Roles are "scaffold" and/or "toolchain".
	Roles []string `toml:"roles"`
	// Needs, Default, Many, Requires and Conflicts are the selection rules the
	// host applies (protocol.Choice).
	Needs   []string `toml:"needs"`
	Default bool     `toml:"default"`
	// Suggested is the choice an asker pre-selects (protocol.Choice).
	Suggested bool     `toml:"suggested"`
	Many      bool     `toml:"many"`
	Requires  []string `toml:"requires"`
	Conflicts []string `toml:"conflicts"`
	// Dependencies join the project's package.json; DevDependencies join the
	// workspace root's. Both name packages: the versions come from the catalog.
	Dependencies    []string `toml:"dependencies"`
	DevDependencies []string `toml:"dev_dependencies"`
	// Overrides join the workspace root's npm overrides, by package name, with
	// the version of the catalog.
	Overrides []string `toml:"overrides"`
	// Tasks are the project's package.json scripts, by task name.
	Tasks map[string]string `toml:"tasks"`
	// Slots are lists of text that the templates of the shared base and of the
	// other choices read: the Vite plugins, the entry imports, the tsconfig
	// include list. A slot grows by each choice that sets it, in selection order.
	Slots map[string][]string `toml:"slots"`
	// Files are Go templates (delimiters [[ ]]) by project-relative path. A
	// choice may replace a file of the base or of a choice it requires.
	Files map[string]string `toml:"files"`
}

// CatalogFile is catalog.toml: the version of every package a choice names.
type CatalogFile struct {
	// Versions maps a package name to the range written in package.json.
	Versions map[string]string `toml:"versions"`
}

// StacksFile is stacks.toml: the stacks the plugin publishes, each a saved
// selection of choices.
type StacksFile struct {
	Stack []StackEntry `toml:"stack"`
}

// StackEntry is one published stack.
type StackEntry struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Choices     []string `toml:"choices"`
	// Toolchain are plugins to add after those the choices imply.
	Toolchain []string `toml:"toolchain"`
}
