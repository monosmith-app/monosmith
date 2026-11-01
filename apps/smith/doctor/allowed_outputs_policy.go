package doctor

import (
	"path"
	"strings"
)

// InsideAny reports whether the workspace-relative path is one of the allowed
// paths or lies under one. Both are cleaned first, so "a/./b/" equals "a/b".
//
// Parameters: p is the path to test; allowed are the paths [workspace]
// allowed_outputs or shared_outputs lists.
// Returns: true when p is an allowed path or under one.
func InsideAny(p string, allowed []string) bool {
	p = path.Clean(p)
	for _, a := range allowed {
		a = path.Clean(a)
		if p == a || strings.HasPrefix(p, a+"/") {
			return true
		}
	}
	return false
}
