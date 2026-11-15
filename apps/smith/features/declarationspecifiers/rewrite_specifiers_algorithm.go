package declarationspecifiers

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	specifier = regexp.MustCompile(`((?:from|import)\s*(?:\(\s*)?)(['"])(\.{1,2}(?:/[^'"]*)?)(['"])`)
	resolved  = regexp.MustCompile(`\.(?:[cm]?js|json)$`)
)

// Rewrite gives every relative specifier in the declaration text the file it
// names, as Node resolves ES modules: `./hooks` is `./hooks.js` when
// hooks.d.ts sits beside the file, and `./hooks/index.js` when hooks is a
// folder with an index.d.ts. dir is the folder of the file the text is from,
// and exists says whether a path exists. A specifier that names nothing found
// is left as it is.
func Rewrite(text, dir string, exists func(path string) bool) string {
	return specifier.ReplaceAllStringFunc(text, func(match string) string {
		m := specifier.FindStringSubmatch(match)
		prefix, open, name, closing := m[1], m[2], m[3], m[4]
		if open != closing || resolved.MatchString(name) {
			return match
		}
		bare := strings.TrimSuffix(name, "/")
		if exists(filepath.Join(dir, bare+".d.ts")) {
			return prefix + open + bare + ".js" + open
		}
		if exists(filepath.Join(dir, bare, "index.d.ts")) {
			return prefix + open + bare + "/index.js" + open
		}
		return match
	})
}
