package imports

import (
	"path"
	"regexp"
	"strings"
)

var (
	pyFrom   = regexp.MustCompile(`(?m)^[ \t]*from[ \t]+(\.*)([\w.]*)[ \t]+import[ \t]+([^\n#]*)`)
	pyImport = regexp.MustCompile(`(?m)^[ \t]*import[ \t]+([\w., \t]+)`)
)

var pythonLanguage = Language{
	Name:       "python",
	Extensions: []string{".py"},
	Inputs:     []string{"**/*.py"},
	IsTest: func(name string) bool {
		return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test.py") || name == "conftest.py"
	},
	Extract: pythonImports,
}

// pythonImports reads "from x import y" and "import x". A relative import is
// resolved from the file's folder; an absolute one belongs to the project when
// its first name is a top-level package or module of the project.
func pythonImports(p Project, rel, text string) []Ref {
	var refs []Ref
	dir := split(path.Dir(rel), "/")
	if path.Dir(rel) == "." {
		dir = nil
	}
	for _, m := range pyFrom.FindAllStringSubmatch(text, -1) {
		dots, module, names := len(m[1]), split(m[2], "."), m[3]
		spec := m[1] + m[2]
		if dots > 0 {
			up := dots - 1
			if up > len(dir) {
				continue
			}
			base := append(append([]string(nil), dir[:len(dir)-up]...), module...)
			if len(module) == 0 {
				for _, name := range importedNames(names) {
					refs = append(refs, Ref{Spec: spec + " import " + name, Segments: append(append([]string(nil), base...), name)})
				}
				continue
			}
			refs = append(refs, Ref{Spec: spec, Segments: base})
			continue
		}
		if len(module) > 0 && p.PythonRoots[module[0]] {
			refs = append(refs, Ref{Spec: spec, Segments: module})
		}
	}
	for _, m := range pyImport.FindAllStringSubmatch(text, -1) {
		for _, item := range strings.Split(m[1], ",") {
			fields := strings.Fields(item)
			if len(fields) == 0 {
				continue
			}
			module := split(fields[0], ".")
			if len(module) > 0 && p.PythonRoots[module[0]] {
				refs = append(refs, Ref{Spec: fields[0], Segments: module})
			}
		}
	}
	return refs
}

// importedNames lists the names after "import" on one line.
func importedNames(list string) []string {
	var out []string
	for _, item := range strings.Split(strings.NewReplacer("(", " ", ")", " ", "\\", " ").Replace(list), ",") {
		if fields := strings.Fields(item); len(fields) > 0 && fields[0] != "*" {
			out = append(out, fields[0])
		}
	}
	return out
}
