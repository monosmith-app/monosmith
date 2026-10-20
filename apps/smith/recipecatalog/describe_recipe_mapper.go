package recipecatalog

import (
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// Description is a recipe as written, before any choice is answered: every
// choice it can ask (also inside when blocks), every level or slot it fills
// and the plugins it configures.
type Description struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	// Chain is the recipes involved, the oldest ancestor first.
	Chain   []recipe.Link     `json:"chain"`
	Choices []DescribedChoice `json:"choices"`
	// Uses are the levels and slots the chain fills, with every plugin it may
	// put there depending on the answers.
	Uses []DescribedUse `json:"use"`
	// Configures names the plugins whose settings the chain sets.
	Configures []string `json:"configures"`
}

// DescribedChoice is a choice and where it is asked.
type DescribedChoice struct {
	Name    string   `json:"name"`
	Ask     string   `json:"ask,omitempty"`
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
	// When is "<choice>=<answer>" when the choice is asked only inside that
	// when block (nested blocks are joined by a comma); empty when it is
	// always asked.
	When   string        `json:"when,omitempty"`
	Origin recipe.Origin `json:"origin"`
}

// DescribedUse is a level or slot: the plugins the recipe puts there always
// (Plugins) or only for an answer (Alternatives, by "<choice>=<answer>").
type DescribedUse struct {
	Key          string              `json:"key"`
	Plugins      []string            `json:"plugins,omitempty"`
	Alternatives map[string][]string `json:"alternatives,omitempty"`
}

// Describe reads what a loaded recipe offers without answering its choices.
func Describe(loaded recipe.Loaded) Description {
	d := Description{Chain: []recipe.Link{}, Choices: []DescribedChoice{}, Uses: []DescribedUse{}, Configures: []string{}}
	uses := map[string]*DescribedUse{}
	configures := map[string]bool{}
	use := func(key string) *DescribedUse {
		if uses[key] == nil {
			uses[key] = &DescribedUse{Key: key}
		}
		return uses[key]
	}
	var walk func(b recipe.Block, doc recipe.Document, when, path string)
	walk = func(b recipe.Block, doc recipe.Document, when, path string) {
		for _, name := range sortedKeys(b.Choices) {
			c := b.Choices[name]
			d.Choices = append(d.Choices, DescribedChoice{
				Name: name, Ask: c.Ask, Options: c.Options, Default: c.Default, When: when,
				Origin: recipe.Origin{Recipe: doc.Recipe.Name, File: doc.Source.Ref, Line: doc.Line(path + "choices." + name)},
			})
		}
		for _, key := range sortedKeys(b.Use) {
			plugins := pluginList(b.Use[key])
			u := use(key)
			if when == "" {
				u.Plugins = plugins
				continue
			}
			if u.Alternatives == nil {
				u.Alternatives = map[string][]string{}
			}
			u.Alternatives[when] = plugins
		}
		for plugin := range b.Settings {
			configures[plugin] = true
		}
		for _, choice := range sortedKeys(b.When) {
			for _, answer := range sortedKeys(b.When[choice]) {
				label := choice + "=" + answer
				if when != "" {
					label = when + "," + label
				}
				walk(b.When[choice][answer], doc, label, path+"when."+choice+"."+answer+".")
			}
		}
	}
	for _, doc := range loaded.Chain {
		d.Chain = append(d.Chain, recipe.Link{Name: doc.Recipe.Name, Version: doc.Recipe.Version, Source: doc.Source.Ref})
		walk(doc.Recipe.Block, doc, "", "")
		d.Name, d.Version = doc.Recipe.Name, doc.Recipe.Version
	}
	for _, key := range sortedKeys(uses) {
		d.Uses = append(d.Uses, *uses[key])
	}
	for plugin := range configures {
		d.Configures = append(d.Configures, plugin)
	}
	sort.Strings(d.Configures)
	return d
}

// pluginList reads a use value: one plugin, or a list of them.
func pluginList(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// sortedKeys is the keys of a map in order, so a description is stable.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
