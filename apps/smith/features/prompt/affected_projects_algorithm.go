package prompt

import (
	"sort"
	"strings"
)

// AffectedBy names the projects that changed files touch, with every project
// that depends on them, sorted. files are workspace-relative with forward
// slashes. A file outside every project root touches nothing here: the prompt
// does not guess at workspace-level files the way smith affected does.
func AffectedBy(projects []Project, files []string) []string {
	byName := map[string]Project{}
	for _, p := range projects {
		byName[p.Name] = p
	}
	hit := map[string]bool{}
	for _, f := range files {
		var best *Project
		for i := range projects {
			p := &projects[i]
			if f == p.Root || strings.HasPrefix(f, p.Root+"/") {
				if best == nil || len(p.Root) > len(best.Root) {
					best = p
				}
			}
		}
		if best == nil {
			continue
		}
		hit[best.Name] = true
		for _, d := range byName[best.Name].Dependents {
			hit[d] = true
		}
	}
	out := make([]string, 0, len(hit))
	for n := range hit {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
