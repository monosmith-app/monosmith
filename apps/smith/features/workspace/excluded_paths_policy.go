package workspace

import (
	"path"

	"github.com/bmatcuk/doublestar/v4"
)

// Excluded reports whether a workspace-relative file lies in a directory
// that [workspace] exclude leaves out: a glob matches the file's directory
// or one of its parents, so "examples/*" covers examples/movies/package.json
// and everything below examples/movies.
func (c Config) Excluded(file string) bool {
	if len(c.Workspace.Exclude) == 0 {
		return false
	}
	for dir := path.Dir(file); dir != "." && dir != "/"; dir = path.Dir(dir) {
		for _, glob := range c.Workspace.Exclude {
			if ok, _ := doublestar.Match(glob, dir); ok {
				return true
			}
		}
	}
	return false
}
