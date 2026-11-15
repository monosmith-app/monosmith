package recipe

// Recipe is a recipe file as written.
type Recipe struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	// Extends is the parent recipe: a path (relative to this file) or an
	// https URL.
	Extends string `toml:"extends"`
	Block
}

// Block is what a recipe, or one of its when blocks, says. Blocks nest.
type Block struct {
	// Choices are asked of whoever uses the recipe; a when block's own
	// choices are asked only when the block applies.
	Choices map[string]Choice `toml:"choices"`
	// Use fills a level (language, scaffold, structure, deploy) or a slot
	// (the task a toolchain provides: lint, test, ...) with a plugin: a
	// string, or a list of strings for the additive levels.
	Use map[string]any `toml:"use"`
	// Settings configure a plugin, by its name.
	Settings map[string]map[string]any `toml:"settings"`
	// When applies a block only for one answer of one choice:
	// when.<choice>.<answer>.
	When map[string]map[string]Block `toml:"when"`
}

// Choice is a question a recipe asks.
type Choice struct {
	Ask     string   `toml:"ask"`
	Options []string `toml:"options"`
	Default string   `toml:"default"`
}

// Source says where a recipe came from.
type Source struct {
	// Ref is the path or URL it was loaded from.
	Ref string
	// URL reports whether Ref is an https URL.
	URL bool
}

// Document is a parsed recipe with the lines its keys are on.
type Document struct {
	Recipe Recipe
	Source Source
	lines  map[string]int
}

// Line is the line of a key, given by its dotted path from the top of the
// file (use.language, settings.eslint.config, when.language.python), or 0
// when it is unknown. A table's line is the first line of any key under it.
func (d Document) Line(path string) int { return d.lines[path] }
