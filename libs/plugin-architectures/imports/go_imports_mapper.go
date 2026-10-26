package imports

import (
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

var goLanguage = Language{
	Name:       "go",
	Extensions: []string{".go"},
	Inputs:     []string{"**/*.go"},
	IsTest:     func(name string) bool { return strings.HasSuffix(name, "_test.go") },
	Extract:    goImports,
}

// goImports reads the import block with Go's own parser. An import of the
// project is one under the module path of go.mod.
func goImports(p Project, _, text string) []Ref {
	if p.GoModule == "" {
		return nil
	}
	file, err := parser.ParseFile(token.NewFileSet(), "", text, parser.ImportsOnly)
	if file == nil || (err != nil && len(file.Imports) == 0) {
		return nil
	}
	var refs []Ref
	for _, spec := range file.Imports {
		target, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if target == p.GoModule {
			refs = append(refs, Ref{Spec: target})
		} else if rest, ok := strings.CutPrefix(target, p.GoModule+"/"); ok {
			refs = append(refs, Ref{Spec: target, Segments: split(rest, "/")})
		}
	}
	return refs
}
