package pruning

import (
	"encoding/json"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// installSections are the package.json sections a production install reads.
var installSections = []string{"dependencies", "optionalDependencies", "peerDependencies"}

// PrunedManifest rewrites a package.json for an install outside the
// workspace. Each workspace package it uses (isWorkspace) points at its copy
// (spec(name), a file: path); a workspace peer becomes a dependency, since a
// file: spec is not a peer's to have. devDependencies and workspaces go: the
// deploy installs the production tree alone. It returns the new file and
// the workspace packages it uses, sorted.
func PrunedManifest(raw []byte, isWorkspace func(string) bool, spec func(name string) string) ([]byte, []string, error) {
	var sections map[string]json.RawMessage
	if err := json.Unmarshal(raw, &sections); err != nil {
		return nil, nil, err
	}
	deps := map[string]map[string]string{}
	for _, s := range installSections {
		m := map[string]string{}
		if section, ok := sections[s]; ok {
			if err := json.Unmarshal(section, &m); err != nil {
				return nil, nil, err
			}
		}
		deps[s] = m
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil, nil, err
	}
	used := map[string]bool{}
	for _, s := range installSections {
		for name := range deps[s] {
			if !isWorkspace(name) {
				continue
			}
			used[name] = true
			if s == "peerDependencies" {
				delete(deps[s], name)
				deps["dependencies"][name] = spec(name)
				continue
			}
			deps[s][name] = spec(name)
		}
	}
	for _, s := range installSections {
		if _, had := sections[s]; !had && len(deps[s]) == 0 {
			continue
		}
		if len(deps[s]) == 0 {
			doc.Delete(s)
			continue
		}
		if err := doc.SetJSON(s, deps[s]); err != nil {
			return nil, nil, err
		}
	}
	if meta := doc.Object("peerDependenciesMeta", false); meta != nil {
		for name := range used {
			meta.Delete(name)
		}
		if meta.Len() == 0 {
			doc.Delete("peerDependenciesMeta")
		}
	}
	doc.Delete("devDependencies")
	doc.Delete("workspaces")
	names := make([]string, 0, len(used))
	for name := range used {
		names = append(names, name)
	}
	sort.Strings(names)
	return doc.Bytes(), names, nil
}
