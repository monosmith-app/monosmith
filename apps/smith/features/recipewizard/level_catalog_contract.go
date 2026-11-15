package recipewizard

// PluginOffer is one plugin a level can take, as `smith recipe new
// --list-levels --json` prints it.
type PluginOffer struct {
	// Name is what goes after "level=" in --fixed.
	Name string `json:"name"`
	// Description says the levels the plugin plays and what it covers.
	Description string `json:"description"`
}

// LevelOffer is one level (language, scaffold, a task slot, structure,
// deploy, toolchain) and every plugin that can fill it.
type LevelOffer struct {
	Level string `json:"level"`
	// Additive is true when the level takes any number of plugins.
	Additive bool          `json:"additive"`
	Plugins  []PluginOffer `json:"plugins"`
}

// LevelCatalog is the one source of the plugins a recipe form offers: the
// terminal wizard, the interactive screen and the editor extensions read it.
type LevelCatalog struct {
	// Levels are the levels some plugin can fill, in the order a recipe is
	// written in.
	Levels []LevelOffer `json:"levels"`
	// Fits says, for each language plugin, the plugins of each other level that
	// fit that language: Fits[language][level] is a list of plugin names. A
	// level a language has no plugin for is absent.
	Fits map[string]map[string][]string `json:"fits"`
}
