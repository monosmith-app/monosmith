package verifyfacts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Build is what a project's build wrote.
type Build struct {
	// Path is the folder or file, workspace-relative with forward slashes.
	Path string
	// Bytes is the size of everything under it.
	Bytes int64
}

// FindBuild measures what a build wrote.
//
// Parameters: workspaceRoot is the workspace folder; output is the build output,
// workspace-relative (protocol.BuildOutputDir of the project).
// Returns: the path and the size, and true when the output exists.
func FindBuild(workspaceRoot, output string) (Build, bool) {
	root := filepath.Join(workspaceRoot, filepath.FromSlash(output))
	info, err := os.Stat(root)
	if err != nil {
		return Build{}, false
	}
	size := info.Size()
	if info.IsDir() {
		size = 0
		_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if fi, err := d.Info(); err == nil {
				size += fi.Size()
			}
			return nil
		})
	}
	return Build{Path: filepath.ToSlash(filepath.Clean(output)), Bytes: size}, true
}

// FormatSize words a size as "980 B", "12.4 KB", "4.2 MB" or "1.3 GB".
func FormatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n), []string{"KB", "MB", "GB", "TB"}
	i := -1
	for value >= unit && i < len(suffix)-1 {
		value /= unit
		i++
	}
	return fmt.Sprintf("%.1f %s", value, suffix[i])
}
