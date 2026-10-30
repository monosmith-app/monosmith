package discovery

import (
	"os"
	"path"
	"path/filepath"
)

// isAppShell reports whether root is the src-tauri folder of a Node project:
// the Rust half of a desktop app, owned by that project and not a crate of the
// monorepo.
func isAppShell(workspaceRoot, root string) bool {
	if path.Base(root) != "src-tauri" {
		return false
	}
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(path.Dir(root)), "package.json"))
	return err == nil
}
