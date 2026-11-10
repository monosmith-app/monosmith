package choicedata

import (
	"fmt"
	"slices"
)

// Plan is what a selection of choices adds to a new project, merged: the
// base first, then each chosen choice of the set in the order given.
type Plan struct {
	// Applied are the names of the choices of the set the plan holds.
	Applied []string
	// Dependencies go into the project's package.json, with the catalog's versions.
	Dependencies map[string]string
	// DevDependencies go into the workspace root's package.json.
	DevDependencies map[string]string
	// Overrides go into the workspace root's npm overrides.
	Overrides map[string]string
	// Tasks are the project's scripts.
	Tasks map[string]string
	// Slots hold the text for the templates, grown by each layer in order.
	Slots map[string][]string
	// Files are the templates, by project-relative path.
	Files map[string]string
}

// Plan merges the base and the named choices that the set owns; names of
// choices of other plugins are skipped. A task or a file that two layers
// write is allowed only when the later layer is the base's or requires the
// earlier one (chakra requires react, so chakra may replace react's main.tsx);
// otherwise it is an error that names both, because the result would depend
// on the order of the names. Slots only grow.
//
// It returns an error for such a clash.
func (s *Set) Plan(names []string) (Plan, error) {
	p := Plan{Dependencies: map[string]string{}, DevDependencies: map[string]string{}, Overrides: map[string]string{},
		Tasks: map[string]string{}, Slots: map[string][]string{}, Files: map[string]string{}}
	var layers []choice
	if s.base != nil {
		base := *s.base
		base.Name = "base"
		layers = append(layers, base)
	}
	for _, name := range names {
		if i := slices.IndexFunc(s.choices, func(c choice) bool { return c.Name == name }); i >= 0 {
			layers = append(layers, s.choices[i])
			p.Applied = append(p.Applied, name)
		}
	}
	owner := map[string]string{}
	for _, l := range layers {
		for _, kind := range []struct {
			what   string
			from   map[string]string
			into   map[string]string
			keyPfx string
		}{{"file", l.Files, p.Files, "file "}, {"task", l.Tasks, p.Tasks, "task "}} {
			for key, value := range kind.from {
				id := kind.keyPfx + key
				if prev, taken := owner[id]; taken && prev != "base" && !s.requires(l.Name, prev) {
					return Plan{}, fmt.Errorf("%s and %s both set the %s %s, and %s does not require %s", prev, l.Name, kind.what, key, l.Name, prev)
				}
				owner[id] = l.Name
				kind.into[key] = value
			}
		}
		for slot, items := range l.Slots {
			p.Slots[slot] = append(p.Slots[slot], items...)
		}
		for _, d := range []struct {
			names []string
			into  map[string]string
		}{{l.Dependencies, p.Dependencies}, {l.DevDependencies, p.DevDependencies}, {l.Overrides, p.Overrides}} {
			for _, pkg := range d.names {
				d.into[pkg] = s.versions[pkg]
			}
		}
	}
	return p, nil
}

// requires reports whether the choice a requires the choice b, directly or
// through other choices of the set.
func (s *Set) requires(a, b string) bool {
	seen := map[string]bool{}
	var walk func(name string) bool
	walk = func(name string) bool {
		i := slices.IndexFunc(s.choices, func(c choice) bool { return c.Name == name })
		if i < 0 || seen[name] {
			return false
		}
		seen[name] = true
		for _, r := range s.choices[i].Requires {
			if r == b || walk(r) {
				return true
			}
		}
		return false
	}
	return walk(a)
}
