package toolrun

import (
	"os"
	"path/filepath"
)

// ManagedBinDir is the folder where smith keeps the tool versions it installed
// for a workspace, at the version the workspace pins (golangci-lint at the
// version CI installs). It is inside .smith, which git ignores, so the copy is
// per machine and per workspace, and it never touches the system's own copy.
// Tasks put it first on PATH, and the checks that probe the tool look there
// first, so both see the same program whatever PATH says.
//
// Parameters: workspaceRoot is the absolute workspace root.
// Returns: the folder, which may not exist yet.
func ManagedBinDir(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".smith", "bin")
}

// ManagedBinExists reports whether smith installed anything into the managed folder.
func ManagedBinExists(workspaceRoot string) bool {
	info, err := os.Stat(ManagedBinDir(workspaceRoot))
	return err == nil && info.IsDir()
}
