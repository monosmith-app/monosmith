package pluginmarket

import (
	"sort"
	"strings"
)

// Search lists the plugins whose name, description, roles, kinds or languages
// contain every word of term, ignoring case; an empty term lists all of them.
// The result is sorted by name.
func Search(idx Index, term string) []Plugin {
	words := strings.Fields(strings.ToLower(term))
	var out []Plugin
	for _, p := range idx.Plugins {
		fields := append([]string{p.Name, p.Description}, p.Roles...)
		fields = append(append(fields, p.Kinds...), p.Languages...)
		hay := strings.ToLower(strings.Join(fields, " "))
		match := true
		for _, w := range words {
			if !strings.Contains(hay, w) {
				match = false
				break
			}
		}
		if match {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
