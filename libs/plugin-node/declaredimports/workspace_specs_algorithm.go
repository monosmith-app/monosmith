package declaredimports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SpecUse is a version range some of the workspace declares a package with.
type SpecUse struct {
	Spec string
	// Projects is how many package.json files declare it with this range.
	Projects int
}

// WorkspaceSpecs lists, for each package, the ranges the root and the
// workspace members declare it with, the most used first (ties by range), in
// any section. A range npm cannot install from the registry (workspace:,
// file:, link:) is left out. It is where a missing declaration takes its
// version from, so the workspace keeps one version of a package.
func WorkspaceSpecs(workspaceRoot string) map[string][]SpecUse {
	counts := map[string]map[string]int{}
	add := func(file string) {
		pkg, err := readManifest(file)
		if err != nil {
			return
		}
		for _, section := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies, pkg.OptionalDependencies} {
			seen := map[string]bool{}
			for name, spec := range section {
				if strings.HasPrefix(spec, "workspace:") || strings.HasPrefix(spec, "file:") || strings.HasPrefix(spec, "link:") || seen[name] {
					continue
				}
				seen[name] = true
				if counts[name] == nil {
					counts[name] = map[string]int{}
				}
				counts[name][spec]++
			}
		}
	}
	add(filepath.Join(workspaceRoot, "package.json"))
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json")); err == nil {
		var root struct {
			Workspaces json.RawMessage `json:"workspaces"`
		}
		if json.Unmarshal(raw, &root) == nil {
			for _, pattern := range workspacePatterns(root.Workspaces) {
				matches, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(pattern), "package.json"))
				for _, file := range matches {
					add(file)
				}
			}
		}
	}
	out := map[string][]SpecUse{}
	for name, specs := range counts {
		for spec, n := range specs {
			out[name] = append(out[name], SpecUse{Spec: spec, Projects: n})
		}
		list := out[name]
		sort.Slice(list, func(i, j int) bool {
			if list[i].Projects != list[j].Projects {
				return list[i].Projects > list[j].Projects
			}
			return list[i].Spec < list[j].Spec
		})
	}
	return out
}

// InstalledVersion is the version of a package installed at the workspace
// root's node_modules, "" when it is not.
func InstalledVersion(workspaceRoot, name string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "node_modules", filepath.FromSlash(name), "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct{ Version string }
	_ = json.Unmarshal(raw, &pkg)
	return pkg.Version
}
