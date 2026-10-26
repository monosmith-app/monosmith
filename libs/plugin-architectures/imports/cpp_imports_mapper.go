package imports

import (
	"regexp"
	"strings"
)

var cppInclude = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*include[ \t]+"([^"\n]+)"`)

var cppLanguage = Language{
	Name:       "cpp",
	Extensions: []string{".cpp", ".cc", ".cxx", ".c", ".h", ".hpp", ".hh"},
	Inputs:     []string{"**/*.cpp", "**/*.cc", "**/*.cxx", "**/*.c", "**/*.h", "**/*.hpp", "**/*.hh"},
	IsTest: func(name string) bool {
		base := name[:len(name)-len(extOf(name))]
		return strings.HasSuffix(base, "_test") || strings.HasSuffix(base, "Test")
	},
	Extract: func(_ Project, rel, text string) []Ref {
		var refs []Ref
		for _, m := range cppInclude.FindAllStringSubmatch(text, -1) {
			spec := m[1]
			parts := split(spec, "/")
			if strings.HasPrefix(spec, ".") {
				var ok bool
				if parts, ok = relative(rel, spec); !ok {
					continue
				}
			}
			refs = append(refs, Ref{Spec: spec, Segments: stripExt(parts)})
		}
		return refs
	},
}

func extOf(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i:]
	}
	return ""
}
