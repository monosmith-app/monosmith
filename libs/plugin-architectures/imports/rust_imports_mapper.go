package imports

import (
	"regexp"
	"strings"
)

var rustUse = regexp.MustCompile(`(?m)^[ \t]*(?:pub(?:\([\w ]+\))?[ \t]+)?use[ \t]+((?:crate|super|self)(?:::\w+)*)`)

var rustLanguage = Language{
	Name:       "rust",
	Extensions: []string{".rs"},
	Inputs:     []string{"src/**/*.rs"},
	IsTest:     func(name string) bool { return strings.HasSuffix(name, "_test.rs") },
	Extract:    rustImports,
}

// rustImports reads use declarations that start with crate, super or self.
// The module path of the file itself is its folders under src/, plus its name
// unless it is mod.rs, lib.rs or main.rs.
func rustImports(_ Project, rel, text string) []Ref {
	here := rustModule(rel)
	var refs []Ref
	for _, m := range rustUse.FindAllStringSubmatch(text, -1) {
		parts := strings.Split(m[1], "::")
		base, next := here, 1
		switch parts[0] {
		case "crate":
			base = nil
		case "super":
			next = 0
			for next < len(parts) && parts[next] == "super" && len(base) > 0 {
				base = base[:len(base)-1]
				next++
			}
		}
		refs = append(refs, Ref{Spec: m[1], Segments: append(append([]string(nil), base...), parts[next:]...)})
	}
	return refs
}

func rustModule(rel string) []string {
	parts := split(rel, "/")
	if len(parts) > 0 && parts[0] == "src" {
		parts = parts[1:]
	}
	parts = stripExt(parts)
	if n := len(parts); n > 0 {
		switch parts[n-1] {
		case "mod", "lib", "main":
			parts = parts[:n-1]
		}
	}
	return parts
}
