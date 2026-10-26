package imports

import (
	"regexp"
	"strings"
)

var dartSpec = regexp.MustCompile(`(?m)^[ \t]*(?:import|export)[ \t]+['"]([^'"\n]+)['"]`)

var dartLanguage = Language{
	Name:       "dart",
	Extensions: []string{".dart"},
	Inputs:     []string{"lib/**/*.dart", "bin/**/*.dart"},
	IsTest:     func(name string) bool { return strings.HasSuffix(name, "_test.dart") },
	Extract:    dartImports,
}

// dartImports reads import and export directives. A package: import of the
// project's own package and a relative path are the project's own code.
func dartImports(p Project, rel, text string) []Ref {
	var refs []Ref
	for _, m := range dartSpec.FindAllStringSubmatch(text, -1) {
		spec := m[1]
		switch {
		case strings.HasPrefix(spec, "dart:"):
		case strings.HasPrefix(spec, "package:"):
			if rest, ok := strings.CutPrefix(spec, "package:"+p.DartPackage+"/"); ok && p.DartPackage != "" {
				refs = append(refs, Ref{Spec: spec, Segments: stripExt(split(rest, "/"))})
			}
		default:
			if parts, ok := relative(rel, spec); ok {
				refs = append(refs, Ref{Spec: spec, Segments: stripExt(parts)})
			}
		}
	}
	return refs
}
