package pluginmarket

// SchemaVersion is the only index format this smith reads.
const SchemaVersion = 1

// Index is the marketplace file: the plugins it lists, by name.
type Index struct {
	// Schema is the format version; it must be SchemaVersion.
	Schema int `json:"schema"`
	// Plugins maps a plugin name to its entry.
	Plugins map[string]Plugin `json:"plugins"`
}

// Plugin is one listed plugin.
type Plugin struct {
	// Name is the key in the index; ParseIndex fills it.
	Name        string `json:"-"`
	Description string `json:"description"`
	// Roles, Kinds and Languages say what the plugin does, for search.
	Roles     []string `json:"roles,omitempty"`
	Kinds     []string `json:"kinds,omitempty"`
	Languages []string `json:"languages,omitempty"`
	Homepage  string   `json:"homepage,omitempty"`
	License   string   `json:"license"`
	// Versions maps a version to the archive that holds it.
	Versions map[string]Release `json:"versions"`
}

// Release is one version of a plugin.
type Release struct {
	// Version is the key in Versions; ParseIndex fills it.
	Version string `json:"-"`
	// URL is where the archive made by `smith plugin pack` is.
	URL string `json:"url"`
	// SHA256 is the lowercase hex digest of that archive.
	SHA256 string `json:"sha256"`
}
