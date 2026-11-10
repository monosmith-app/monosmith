package adoptsplit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// SortNode reads the package.json at the repository root and sorts its
// devDependencies: the tools go to the workspace, the ESLint stack is
// replaced by the shared lint config, and the others are left for the person
// to place. The dependencies (the ones the code runs on) are not listed: they
// always stay with the project. Nothing is changed.
//
// Parameters: root is the repository root.
// Returns: the sorted names, or an error when package.json cannot be read.
func SortNode(root string) (Sorted, error) {
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return Sorted{}, err
	}
	var pkg struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return Sorted{}, fmt.Errorf("package.json: %w", err)
	}
	var out Sorted
	for name := range pkg.DevDependencies {
		switch {
		case IsNodeLint(name):
			out.Replaced = append(out.Replaced, name)
		case IsNodeTool(name):
			out.Tools = append(out.Tools, name)
		default:
			out.Unsure = append(out.Unsure, name)
		}
	}
	sort.Strings(out.Replaced)
	sort.Strings(out.Tools)
	sort.Strings(out.Unsure)
	return out, nil
}
