package linking

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// members reads the package.json of the root and of each workspace member
// ("workspaces" is a list, or an object with "packages"), as raw bytes.
func members(workspaceRoot string) [][]byte {
	rootFile := filepath.Join(workspaceRoot, "package.json")
	raw, err := os.ReadFile(rootFile)
	if err != nil {
		return nil
	}
	out := [][]byte{raw}
	var root struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if json.Unmarshal(raw, &root) != nil {
		return out
	}
	var patterns []string
	if json.Unmarshal(root.Workspaces, &patterns) != nil {
		var object struct {
			Packages []string `json:"packages"`
		}
		if json.Unmarshal(root.Workspaces, &object) == nil {
			patterns = object.Packages
		}
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(pattern), "package.json"))
		for _, file := range matches {
			if data, err := os.ReadFile(file); err == nil {
				out = append(out, data)
			}
		}
	}
	return out
}

// InternalSpecs lists the distinct ranges the workspace's packages already use
// to depend on each other, sorted. It is how the workspace's habit is read: one
// entry is a convention to follow, none means there is no habit yet, several
// mean the workspace is not consistent and the user chooses.
func InternalSpecs(workspaceRoot string) []string {
	manifests := members(workspaceRoot)
	names := map[string]bool{}
	for _, m := range manifests {
		var pkg struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(m, &pkg) == nil && pkg.Name != "" {
			names[pkg.Name] = true
		}
	}
	seen := map[string]bool{}
	for _, m := range manifests {
		for name := range names {
			if spec, ok := DeclaredSpec(m, name); ok {
				seen[spec] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for spec := range seen {
		out = append(out, spec)
	}
	sort.Strings(out)
	return out
}
