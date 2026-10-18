package recipewizard

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// Plan is what a recipe is made from: it is what the wizard asks for, and
// the file --from-answers reads.
type Plan struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	// Extends is the parent: a path, relative to the recipe file, or an https
	// URL. Empty for none.
	Extends string `toml:"extends"`
	// Choices are the questions the recipe asks, by name. A level may use a
	// choice of the parent recipe, which is then not listed here.
	Choices map[string]recipe.Choice `toml:"choices"`
	// Levels says, for each level or slot, which plugin fills it.
	Levels map[string]LevelPlan `toml:"levels"`
}

// LevelPlan fills one level: with plugins that are fixed, or with the
// plugins an answer of a choice picks. Exactly one of the two is set.
type LevelPlan struct {
	// Fixed are the plugins used whatever the answers: one for the language,
	// the scaffold and a slot, any number for structure and deploy.
	Fixed []string `toml:"fixed"`
	// Choice names the choice that picks the plugins, and Plugins gives them
	// by answer: a plugin name, or a list of names for the additive levels.
	Choice  string         `toml:"choice"`
	Plugins map[string]any `toml:"plugins"`
}

// Level keys of the levels that are not task slots.
const (
	LevelLanguage  = "language"
	LevelScaffold  = "scaffold"
	LevelStructure = "structure"
	LevelDeploy    = "deploy"
	LevelToolchain = "toolchain"
)

// Additive reports whether a level takes any number of plugins, as opposed
// to the language, the scaffold and each slot, which take one.
func Additive(level string) bool {
	return level == LevelStructure || level == LevelDeploy || level == LevelToolchain
}

// OrderedLevels lists the levels a plan fills in the order a recipe is
// written in: the language, the scaffold, the slots by name, then structure
// and deploy.
func OrderedLevels(levels map[string]LevelPlan) []string {
	return OrderLevelKeys(keys(levels))
}

// OrderLevelKeys sorts level keys in that order.
func OrderLevelKeys(in []string) []string {
	rank := func(level string) int {
		switch level {
		case LevelLanguage:
			return 0
		case LevelScaffold:
			return 1
		case LevelStructure:
			return 3
		case LevelDeploy:
			return 4
		case LevelToolchain:
			return 5
		}
		return 2
	}
	out := append([]string(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if rank(out[i]) != rank(out[j]) {
			return rank(out[i]) < rank(out[j])
		}
		return out[i] < out[j]
	})
	return out
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
