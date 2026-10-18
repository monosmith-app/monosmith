package workspace

import (
	"path"
	"strings"
)

// ValidContext reports whether a docker build context setting names a
// directory inside the workspace: "." or a relative path with forward
// slashes, never leaving the workspace.
func ValidContext(c string) bool {
	if c == "." {
		return true
	}
	if strings.ContainsAny(c, `\:`) || strings.HasPrefix(c, "/") || path.Clean(c) != strings.TrimSuffix(c, "/") {
		return false
	}
	return c != ".." && !strings.HasPrefix(c, "../")
}
