package imports

import (
	"regexp"
	"strings"
)

var nodeSpec = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*\(?\s*|\brequire\s*\(\s*)['"]([^'"\n]+)['"]`)

var nodeLanguage = Language{
	Name:       "node",
	Extensions: []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts"},
	Inputs:     []string{"**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx", "**/*.mjs", "**/*.cjs", "**/*.mts", "**/*.cts"},
	IsTest: func(name string) bool {
		return strings.Contains(name, ".test.") || strings.Contains(name, ".spec.")
	},
	Extract: nodeImports,
}

// nodeAliasPrefixes are the path aliases a TypeScript project maps to its
// source root; anything else that is not relative is a package.
var nodeAliasPrefixes = []string{"@/", "~/", "#/", "src/"}

// nodeImports reads import, export-from, dynamic import and require. A
// relative path is resolved from the file's folder; an alias counts as the
// project's own code.
func nodeImports(_ Project, rel, text string) []Ref {
	var refs []Ref
	for _, m := range nodeSpec.FindAllStringSubmatch(text, -1) {
		spec := m[1]
		if strings.HasPrefix(spec, ".") {
			if parts, ok := relative(rel, spec); ok {
				refs = append(refs, Ref{Spec: spec, Segments: stripExt(parts)})
			}
			continue
		}
		for _, prefix := range nodeAliasPrefixes {
			if rest, ok := strings.CutPrefix(spec, prefix); ok {
				refs = append(refs, Ref{Spec: spec, Segments: stripExt(split(rest, "/"))})
				break
			}
		}
	}
	return refs
}
