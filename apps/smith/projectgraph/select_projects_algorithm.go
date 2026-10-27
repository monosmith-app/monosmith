package projectgraph

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

// Select resolves selection patterns to project names, sorted. A pattern is a
// name, a glob over names or roots ("lib-*", "apps/*"), "kind:<app|lib|package>",
// "lang:<language>" or "tag:<tag>"; a leading "!" excludes. Only exclusions
// means "everything except". Patterns may be comma-separated.
func (g *Graph) Select(patterns []string) ([]string, error) {
	var include, exclude []string
	for _, raw := range patterns {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			switch {
			case p == "":
			case strings.HasPrefix(p, "!"):
				exclude = append(exclude, p[1:])
			default:
				include = append(include, p)
			}
		}
	}
	chosen := map[string]bool{}
	if len(include) == 0 {
		for _, p := range g.Projects {
			chosen[p.Name] = true
		}
	}
	for _, pattern := range include {
		matched := false
		for _, p := range g.Projects {
			if matches(p, pattern) {
				chosen[p.Name] = true
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("no project matches %q (projects: %s)", pattern, strings.Join(g.Names(), ", "))
		}
	}
	for _, pattern := range exclude {
		for _, p := range g.Projects {
			if matches(p, pattern) {
				delete(chosen, p.Name)
			}
		}
	}
	var out []string
	for _, p := range g.Projects {
		if chosen[p.Name] {
			out = append(out, p.Name)
		}
	}
	return out, nil
}

func matches(p *Project, pattern string) bool {
	if kind, ok := strings.CutPrefix(pattern, "kind:"); ok {
		return string(p.Kind) == kind
	}
	if lang, ok := strings.CutPrefix(pattern, "lang:"); ok {
		return p.Language == lang || p.Plugin == lang
	}
	if tag, ok := strings.CutPrefix(pattern, "tag:"); ok {
		return slices.Contains(p.Tags, tag)
	}
	if p.Name == pattern || p.Root == pattern {
		return true
	}
	if ok, _ := path.Match(pattern, p.Name); ok {
		return true
	}
	ok, _ := path.Match(pattern, p.Root)
	return ok
}

// Matches reports whether the project is named by a selection pattern (a name,
// a glob, "kind:<kind>", "lang:<language>" or "tag:<tag>").
func (p *Project) Matches(pattern string) bool { return matches(p, pattern) }
