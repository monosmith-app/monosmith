package taskrun

import (
	"os"
	"path/filepath"
	"runtime"
)

var isWindows = runtime.GOOS == "windows"

// Platform is part of every cache key: a binary built on one OS is not the
// output of the same task on another.
var Platform = runtime.GOOS + "/" + runtime.GOARCH

// VenvBin is the executables directory of the workspace's Python
// environment (.venv), or "" when there is none.
func VenvBin(root string) string {
	dir := filepath.Join(root, ".venv", "bin")
	if isWindows {
		dir = filepath.Join(root, ".venv", "Scripts")
	}
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}
