package imports

import (
	"regexp"
	"strings"
)

var rubyRequire = regexp.MustCompile(`(?m)^[ \t]*require_relative[ \t(]+['"]([^'"\n]+)['"]`)

var rubyLanguage = Language{
	Name:       "ruby",
	Extensions: []string{".rb"},
	Inputs:     []string{"lib/**/*.rb", "app/**/*.rb"},
	IsTest: func(name string) bool {
		return strings.HasSuffix(name, "_spec.rb") || strings.HasSuffix(name, "_test.rb")
	},
	Extract: func(_ Project, rel, text string) []Ref {
		var refs []Ref
		for _, m := range rubyRequire.FindAllStringSubmatch(text, -1) {
			if parts, ok := relative(rel, m[1]); ok {
				refs = append(refs, Ref{Spec: m[1], Segments: stripExt(parts)})
			}
		}
		return refs
	},
}
