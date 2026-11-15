package adoptsplit

import (
	"os"
	"path/filepath"
	"slices"
)

// SortPython reads the pyproject.toml at the repository root and sorts the
// entries of its development lists (the dev, test and lint extras and
// groups): the tools go to the workspace's requirements-dev.txt, the others
// are left for the person to place. The runtime dependencies are never
// listed: they stay with the project. Nothing is changed.
//
// Parameters: root is the repository root.
// Returns: the sorted requirement names, or an error when pyproject.toml
// cannot be read.
func SortPython(root string) (Sorted, error) {
	raw, err := os.ReadFile(filepath.Join(root, "pyproject.toml"))
	if err != nil {
		return Sorted{}, err
	}
	var out Sorted
	seen := map[string]bool{}
	for _, g := range devGroups {
		entries, _ := readDevGroup(string(raw), g)
		for _, e := range entries {
			name := RequirementName(e)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			if IsPythonTool(name) {
				out.Tools = append(out.Tools, name)
			} else {
				out.Unsure = append(out.Unsure, name)
			}
		}
	}
	slices.Sort(out.Tools)
	slices.Sort(out.Unsure)
	return out, nil
}
