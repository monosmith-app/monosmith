package declaredimports

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// declaredInWorkspace is every package any workspace member's package.json
// declares, in any section. A test file may import what a sibling declares: the
// workspace installs it, and a test never ships. (Nx's own dependency check let
// tests import anything, so a converted workspace has such imports.)
func declaredInWorkspace(workspaceRoot string) map[string]bool {
	out := map[string]bool{}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json"))
	if err != nil {
		return out
	}
	var root struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if json.Unmarshal(raw, &root) != nil {
		return out
	}
	patterns := workspacePatterns(root.Workspaces)
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(pattern), "package.json"))
		for _, file := range matches {
			pkg, err := readManifest(file)
			if err != nil {
				continue
			}
			for name := range names(pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies, pkg.OptionalDependencies) {
				out[name] = true
			}
		}
	}
	return out
}

// workspacePatterns reads "workspaces" as npm does: a list, or an object with
// a "packages" list.
func workspacePatterns(raw json.RawMessage) []string {
	var list []string
	if json.Unmarshal(raw, &list) == nil {
		return list
	}
	var object struct {
		Packages []string `json:"packages"`
	}
	if json.Unmarshal(raw, &object) == nil {
		return object.Packages
	}
	return nil
}
