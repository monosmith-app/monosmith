package languages

import (
	"regexp"
	"strings"
)

// goPlatforms are the build-constraint suffixes a Go file name may end in.
var goPlatforms = []string{"windows", "linux", "darwin", "freebsd", "js", "wasm", "amd64", "arm64", "386", "arm"}

var (
	platformSuffix = `(?:_(?:` + strings.Join(goPlatforms, "|") + `))*`
	goRoleFile     = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)` + platformSuffix + `(?:_test)?\.go$`)
	goRootFile     = regexp.MustCompile(`^(?:doc|main)` + platformSuffix + `(?:_test)?\.go$`)
)

// goRules: a module's root holds only doc.go and main.go; every other file
// is <snake>_<role>.go.
var goRules = Rules{
	Name:   "go",
	Roots:  []string{""},
	Inputs: []string{"**/*.go"},
	Source: func(name string, _ Dir) bool { return strings.HasSuffix(name, ".go") },
	File:   func(name string, in Dir) string { return GoFile(name, in.Rel == "") },
	Production: func(name string, _ Dir) bool {
		return name != "doc.go" && !strings.HasSuffix(name, "_test.go")
	},
}

// GoFile says why a Go file breaks the rules, or "" when it does not.
func GoFile(name string, atRoot bool) string {
	if atRoot {
		if goRootFile.MatchString(name) {
			return ""
		}
		return "only doc.go and main.go sit at the root of a module; move this into a slice folder"
	}
	if name == "doc.go" || goRoleFile.MatchString(name) {
		return ""
	}
	return "a Go file is <snake>_<role>.go (roles: " + strings.Join(Roles, ", ") + ")"
}
