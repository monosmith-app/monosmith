package declaredimports

import (
	"path"
	"strings"
)

// devFolders hold files that never ship: tests, test doubles, tools and examples.
var devFolders = map[string]bool{"e2e": true, "tools": true, "test": true, "tests": true, "testing": true, "scripts": true, "docs": true, "examples": true, "benchmarks": true}

// DevFile reports whether a file, by its path relative to the project, is a
// test, tool or build script: it may import what the project declares as a
// development dependency, which installing the package does not install.
func DevFile(rel string) bool {
	for _, dir := range strings.Split(path.Dir(rel), "/") {
		if devFolders[dir] {
			return true
		}
	}
	base := path.Base(rel)
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") ||
		base == "conftest.py" || base == "setup.py" || base == "noxfile.py"
}
