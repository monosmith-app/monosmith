package recipedrift

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Value is one thing a recipe decided for a project, and where it said so.
type Value struct {
	// Name is the plugin level ("use.lint"), the setting
	// ("settings.ruff.line") or the choice ("choices.language").
	Name  string `json:"name"`
	Value string `json:"value"`
	// Origin is the recipe, file and line; nil when it has none.
	Origin *recipe.Origin `json:"origin,omitempty"`
}

// Explained is a project's recipe as it resolves now with the answers the
// project recorded.
type Explained struct {
	Project string            `json:"project"`
	Recipe  string            `json:"recipe"`
	Version string            `json:"version,omitempty"`
	Source  string            `json:"source"`
	Answers map[string]string `json:"answers,omitempty"`
	Values  []Value           `json:"values"`
}

// Explain resolves a project's recipe and lists what it decided, each value
// with the recipe, file and line it came from. ok is false for a project that
// records no recipe.
func Explain(s *session.Session, name string) (e Explained, ok bool, err error) {
	ov, found := s.Config.Projects[name]
	p := s.Graph.Get(name)
	if !found || ov.Recipe == nil || p == nil {
		return Explained{}, false, nil
	}
	rec := *ov.Recipe
	loaded, ref, problem := load(s, &recipe.Loader{Root: s.Root}, p, rec)
	if problem != "" {
		return Explained{}, true, fmt.Errorf("%s: %s", name, problem)
	}
	res, err := recipe.Resolve(loaded, rec.Answers)
	if err != nil {
		return Explained{}, true, fmt.Errorf("%s: %w", name, err)
	}
	e = Explained{Project: name, Recipe: res.Name, Version: res.Version, Source: ref, Answers: rec.Answers}
	e.Values = values(res)
	return e, true, nil
}

// values lists a resolution's choices, plugins and settings, in that order,
// each group sorted.
func values(res recipe.Resolution) []Value {
	var out []Value
	origin := func(path string, fallback recipe.Origin) *recipe.Origin {
		if o, ok := res.Origins[path]; ok {
			return &o
		}
		if fallback.File != "" {
			return &fallback
		}
		return nil
	}
	for _, q := range res.Asked {
		if q.Answer != "" {
			o := q.Origin
			out = append(out, Value{Name: "choices." + q.Name, Value: q.Answer, Origin: &o})
		}
	}
	for _, u := range res.Uses {
		out = append(out, Value{Name: "use." + u.Key, Value: strings.Join(u.Plugins, ", "), Origin: origin("use."+u.Key, u.Origin)})
	}
	settings := map[string]string{}
	for plugin, table := range res.Settings {
		if m, ok := table.(map[string]any); ok {
			for path, v := range flatten(m) {
				settings["settings."+plugin+"."+path] = v
			}
		}
	}
	names := make([]string, 0, len(settings))
	for name := range settings {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		out = append(out, Value{Name: name, Value: settings[name], Origin: origin(name, recipe.Origin{})})
	}
	return out
}
