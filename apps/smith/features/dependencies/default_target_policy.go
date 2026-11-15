package dependencies

import (
	"path/filepath"
	"strings"
)

// DefaultTarget is where a package goes when -p is not given: the project
// whose folder holds the current folder (the deepest one, for a project inside
// another), else the workspace root (RootTarget).
//
// Parameters: workspaceRoot and cwd are absolute folders; roots maps each
// project's name to its workspace-relative folder.
// Returns: a project name, or RootTarget.
func DefaultTarget(workspaceRoot, cwd string, roots map[string]string) string {
	rel, err := filepath.Rel(workspaceRoot, cwd)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return RootTarget
	}
	here := filepath.ToSlash(rel) + "/"
	best, depth := RootTarget, 0
	for name, root := range roots {
		root = strings.Trim(root, "/")
		if root == "" || root == "." {
			continue
		}
		if strings.HasPrefix(here, root+"/") && len(root) > depth {
			best, depth = name, len(root)
		}
	}
	return best
}
