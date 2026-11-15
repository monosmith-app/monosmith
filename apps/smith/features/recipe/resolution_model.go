package recipe

import (
	"fmt"
	"sort"
)

// Origin says which recipe, file and line set a value.
type Origin struct {
	Recipe string `json:"recipe"`
	File   string `json:"file"`
	Line   int    `json:"line,omitempty"`
}

func (o Origin) String() string {
	if o.Line == 0 {
		return fmt.Sprintf("%s (recipe %s)", o.File, o.Recipe)
	}
	return fmt.Sprintf("%s:%d (recipe %s)", o.File, o.Line, o.Recipe)
}

// Question is a choice as it was asked, and how it was answered.
type Question struct {
	Name    string   `json:"name"`
	Ask     string   `json:"ask,omitempty"`
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
	// Answer is empty when the choice has not been answered.
	Answer string `json:"answer,omitempty"`
	// Given says the answer was one the caller gave, not the default.
	Given  bool   `json:"given,omitempty"`
	Origin Origin `json:"origin"`
}

// Use is a level or slot and the plugins that fill it. A plugin may carry a
// version constraint after an @ (eslint@^1.2).
type Use struct {
	Key     string   `json:"key"`
	Plugins []string `json:"plugins"`
	Origin  Origin   `json:"origin"`
}

// Link is one recipe in the chain.
type Link struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Source  string `json:"source"`
}

// Resolution is a recipe chain and its answers, resolved: the choices asked,
// the plugin for each level, and the merged settings, each with where it
// came from.
type Resolution struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	// Chain is the recipes involved, the oldest ancestor first.
	Chain    []Link         `json:"chain"`
	Asked    []Question     `json:"choices"`
	Uses     []Use          `json:"use"`
	Settings map[string]any `json:"settings,omitempty"`
	// Origins say where each value came from, by dotted path: use.<key>, and
	// settings.<plugin>.<setting>.
	Origins map[string]Origin `json:"origins,omitempty"`
}

// UseFor returns the entry for a level or slot.
func (r Resolution) UseFor(key string) (Use, bool) {
	for _, u := range r.Uses {
		if u.Key == key {
			return u, true
		}
	}
	return Use{}, false
}

// setUse puts a plugin list on a level, replacing what an earlier recipe
// said, and keeps the entries sorted by key.
func (r *Resolution) setUse(key string, plugins []string, origin Origin) {
	for i := range r.Uses {
		if r.Uses[i].Key == key {
			r.Uses[i] = Use{Key: key, Plugins: plugins, Origin: origin}
			return
		}
	}
	r.Uses = append(r.Uses, Use{Key: key, Plugins: plugins, Origin: origin})
	sort.Slice(r.Uses, func(i, j int) bool { return r.Uses[i].Key < r.Uses[j].Key })
}

// mergeSettings merges a plugin's settings table into what earlier recipes
// set: tables merge key by key, anything else replaces. line finds the line
// of a setting by its path under the plugin.
func (r *Resolution) mergeSettings(plugin string, table map[string]any, recipe, file string, line func(path string) int) {
	if r.Settings == nil {
		r.Settings = map[string]any{}
	}
	if r.Origins == nil {
		r.Origins = map[string]Origin{}
	}
	dst, _ := r.Settings[plugin].(map[string]any)
	if dst == nil {
		dst = map[string]any{}
		r.Settings[plugin] = dst
	}
	mergeTable(dst, table, "settings."+plugin, r.Origins, recipe, file, "", line)
}

func mergeTable(dst, src map[string]any, base string, origins map[string]Origin, recipe, file, rel string, line func(string) int) {
	for _, k := range sortedKeys(src) {
		path := base + "." + k
		relPath := k
		if rel != "" {
			relPath = rel + "." + k
		}
		if table, ok := src[k].(map[string]any); ok {
			into, _ := dst[k].(map[string]any)
			if into == nil {
				into = map[string]any{}
				dst[k] = into
			}
			mergeTable(into, table, path, origins, recipe, file, relPath, line)
			continue
		}
		dst[k] = src[k]
		origins[path] = Origin{Recipe: recipe, File: file, Line: line(relPath)}
	}
}
