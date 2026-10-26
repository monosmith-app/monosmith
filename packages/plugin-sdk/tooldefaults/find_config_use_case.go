package tooldefaults

import (
	"os"
	"path"
	"path/filepath"
)

// FindConfig reports whether the repository configures a tool for a project:
// accept is asked about each directory from the project root up to the
// workspace root, the nearest first, with the directory's absolute path. Tools
// look for their configuration the same way, so the nearest file wins.
func FindConfig(workspaceRoot, projectRoot string, accept func(dir string) bool) bool {
	rel := path.Clean(filepath.ToSlash(projectRoot))
	for {
		dir := workspaceRoot
		if rel != "." {
			dir = filepath.Join(workspaceRoot, filepath.FromSlash(rel))
		}
		if accept(dir) {
			return true
		}
		if rel == "." || rel == "/" {
			return false
		}
		rel = path.Dir(rel)
	}
}

// HasFile is an accept function for FindConfig: the directory holds a file
// with one of the names.
func HasFile(names ...string) func(dir string) bool {
	return func(dir string) bool {
		for _, n := range names {
			if info, err := os.Stat(filepath.Join(dir, n)); err == nil && !info.IsDir() {
				return true
			}
		}
		return false
	}
}
