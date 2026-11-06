package gowork

import (
	"os"
	"path"
	"path/filepath"
)

// ModuleFile is the go.mod of the module a project belongs to: the one in its
// folder, or for a project declared in smith.toml (a package of a module) the
// nearest one in a folder above it. It is the path of the folder's own go.mod
// when none exists above, so a read of it fails as before.
func ModuleFile(workspaceRoot, root string) string {
	for dir := root; ; dir = path.Dir(dir) {
		file := filepath.Join(workspaceRoot, filepath.FromSlash(dir), "go.mod")
		if _, err := os.Stat(file); err == nil {
			return file
		}
		if dir == "." || dir == "/" {
			return filepath.Join(workspaceRoot, filepath.FromSlash(root), "go.mod")
		}
	}
}

// DeclaredInModule says whether a project is a package of a module above its
// folder: it has no go.mod of its own and one exists in a folder above it.
func DeclaredInModule(workspaceRoot, root string) bool {
	own := filepath.Join(workspaceRoot, filepath.FromSlash(root), "go.mod")
	if _, err := os.Stat(own); err == nil {
		return false
	}
	return ModuleFile(workspaceRoot, root) != own
}
