package buildoutputs

import (
	"path"
	"strings"
)

// Covered reports whether a declared output restores dir: it is that
// directory or a parent of it.
func Covered(dir string, declared []string) bool {
	dir = path.Clean(dir)
	for _, d := range declared {
		d = path.Clean(d)
		if d != "." && (dir == d || strings.HasPrefix(dir, d+"/")) {
			return true
		}
	}
	return false
}

// Uncovered returns the written directories (collapsed to their outermost)
// that no declared output restores.
func Uncovered(written, declared []string) []string {
	var out []string
	for _, w := range Collapse(written) {
		if !Covered(w, declared) {
			out = append(out, w)
		}
	}
	return out
}
