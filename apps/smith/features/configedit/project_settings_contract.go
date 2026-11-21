package configedit

// ProjectFacts are what the project's plugin reports about a project, from
// its manifest: the editors show them read-only beside the settings of
// [projects.<name>].
type ProjectFacts struct {
	// Name is the project's name; Root its workspace-relative folder.
	Name, Root string
	// Language is the plugin that describes it; Kind is app, lib or package.
	Language, Kind string
	// Version is the version its manifest records; empty when it has none.
	Version string
}

// ProjectFact is one value a project's manifest or plugin owns, shown
// read-only with the file that holds it.
type ProjectFact struct {
	// Key names it: name, version, language, kind or root.
	Key string `json:"path"`
	// Value is the value as the plugin reports it.
	Value string `json:"value"`
	// Explanation says what it is and where to change it.
	Explanation string `json:"description"`
	// Source is the workspace-relative file that holds it (apps/web/package.json),
	// or the project's folder when smith cannot name the file.
	Source string `json:"source"`
}

// ProjectSetting is one setting of [projects.<name>] and its value now.
type ProjectSetting struct {
	Field
	// Value is the value as written in smith.toml; empty when it is not set.
	Value string `json:"value,omitempty"`
	// Set is false when the setting has its default.
	Set bool `json:"set"`
}

// ProjectView is what the editors show for one project: the values its
// manifest owns, read-only, and the settings smith.toml may give it.
type ProjectView struct {
	// Name is the project's name.
	Name string `json:"name"`
	// Declared is true for a project that smith.toml declares (root and
	// language in [projects.<name>]) because it has no manifest of its own.
	Declared bool `json:"declared"`
	// Manifest is the file that owns the read-only values: the manifest, or
	// smith.toml for a declared project.
	Manifest string `json:"manifest"`
	// Facts are the read-only values.
	Facts []ProjectFact `json:"facts"`
	// Settings are the settings of [projects.<name>] that apply to it: root,
	// language and kind only for a declared project.
	Settings []ProjectSetting `json:"settings"`
}
