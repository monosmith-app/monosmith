package convertplan

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

var rollupRequire = regexp.MustCompile(`require\(\s*['"](\.{1,2}/[^'"]+)['"]\s*\)`)

// rollupDelegate finds the shared file a one-line rollup config builds through
// (module.exports = require('../../rollup.base.cjs')()): when that file is the
// withNx build, it is what the config stands for. ok is false for a config
// that is a build itself or requires something else.
func rollupDelegate(f string, files []string, read func(string) string) (file, text string, ok bool) {
	code := stripLineComments(read(f))
	if strings.Contains(code, "with-nx") {
		return "", "", false
	}
	m := rollupRequire.FindStringSubmatch(code)
	if m == nil {
		return "", "", false
	}
	target := path.Clean(path.Join(path.Dir(f), m[1]))
	for _, candidate := range []string{target, target + ".cjs", target + ".js", target + ".mjs"} {
		if !slices.Contains(files, candidate) {
			continue
		}
		if t := read(candidate); strings.Contains(stripLineComments(t), "with-nx") {
			return candidate, t, true
		}
	}
	return "", "", false
}
