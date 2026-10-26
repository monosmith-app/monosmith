package imports

import (
	"regexp"
	"strings"
)

var (
	jvmPackage = regexp.MustCompile(`(?m)^[ \t]*package[ \t]+([\w.]+)`)
	jvmImport  = regexp.MustCompile(`(?m)^[ \t]*import[ \t]+(?:static[ \t]+)?([\w.]+)`)
)

var jvmLanguage = Language{
	Name:       "jvm",
	Extensions: []string{".java", ".kt", ".kts"},
	Inputs:     []string{"**/*.java", "**/*.kt", "**/*.kts"},
	IsTest: func(name string) bool {
		base := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".java"), ".kts"), ".kt")
		return strings.HasSuffix(base, "Test") || strings.HasSuffix(base, "Tests") || strings.HasSuffix(base, "IT")
	},
	Extract: jvmImports,
}

// jvmImports reads import statements. An import belongs to the project when
// it starts like the file's own package (its first two names: com.acme).
func jvmImports(_ Project, _, text string) []Ref {
	return namespacedImports(text, jvmPackage, jvmImport, ".")
}

// namespacedImports is the shared reading for languages whose files declare a
// namespace and import by qualified name: an import is the project's own when
// it begins with the first names of the file's namespace.
func namespacedImports(text string, namespace, imported *regexp.Regexp, sep string) []Ref {
	ns := namespace.FindStringSubmatch(text)
	if ns == nil {
		return nil
	}
	own := split(ns[1], sep)
	// The last name of a namespace is often the layer itself (Shop.Domain), so
	// it is left out of the comparison, down to its first name.
	depth := max(1, min(2, len(own)-1))
	var refs []Ref
	for _, m := range imported.FindAllStringSubmatch(text, -1) {
		spec := strings.Trim(m[1], sep)
		parts := split(spec, sep)
		if len(parts) < depth || !sameStart(parts, own, depth) {
			continue
		}
		refs = append(refs, Ref{Spec: spec, Segments: parts})
	}
	return refs
}

func sameStart(a, b []string, n int) bool {
	for i := range n {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}
