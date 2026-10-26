package imports

import (
	"path"
	"strings"
)

// split cuts a path or dotted name into its parts, dropping empty ones.
func split(s, sep string) []string {
	var out []string
	for _, part := range strings.Split(s, sep) {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// relative resolves a relative import (./x, ../x) from the folder of the file
// rel. It reports false when the result leaves the project.
func relative(rel, spec string) ([]string, bool) {
	joined := path.Join(path.Dir(rel), spec)
	if joined == ".." || strings.HasPrefix(joined, "../") {
		return nil, false
	}
	return split(joined, "/"), true
}

// stripExt takes a file extension off the last part.
func stripExt(parts []string) []string {
	if len(parts) == 0 {
		return parts
	}
	out := append([]string(nil), parts...)
	last := out[len(out)-1]
	out[len(out)-1] = strings.TrimSuffix(last, path.Ext(last))
	return out
}
