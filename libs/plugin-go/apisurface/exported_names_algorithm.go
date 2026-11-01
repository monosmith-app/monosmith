package apisurface

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strings"
)

// ExportedNames lists the public API of a Go module from its sources: every
// exported function, type, constant and variable, the exported methods of
// exported types, and the exported fields of structs and methods of interfaces.
//
// A name in a package below the module root is prefixed with the package's
// directory and a colon ("client:Dial"). Test files, the packages of an
// "internal", "testdata" or "vendor" directory, hidden directories and package
// main are not part of the API. A file that does not parse is skipped.
//
// Parameters: files maps a file path, relative to the module root and with
// forward slashes, to its text.
// Returns: the names, sorted.
func ExportedNames(files map[string]string) []string {
	seen := map[string]bool{}
	fset := token.NewFileSet()
	for name, text := range files {
		if !isAPIFile(name) {
			continue
		}
		file, err := parser.ParseFile(fset, name, text, parser.SkipObjectResolution)
		if err != nil || file.Name.Name == "main" {
			continue
		}
		prefix := ""
		if dir := path.Dir(name); dir != "." {
			prefix = dir + ":"
		}
		for _, n := range declaredNames(file) {
			seen[prefix+n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// isAPIFile says whether a source file can hold public API.
func isAPIFile(name string) bool {
	if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false
	}
	for _, part := range strings.Split(path.Dir(name), "/") {
		if part == "internal" || part == "testdata" || part == "vendor" || (strings.HasPrefix(part, ".") && part != ".") || strings.HasPrefix(part, "_") {
			return false
		}
	}
	return true
}

// declaredNames are the exported names one file declares.
func declaredNames(file *ast.File) []string {
	var out []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Recv == nil {
				out = append(out, d.Name.Name)
			} else if recv := receiverName(d.Recv); recv != "" {
				out = append(out, recv+"."+d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						out = append(out, s.Name.Name)
						out = append(out, memberNames(s.Name.Name, s.Type)...)
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.IsExported() {
							out = append(out, n.Name)
						}
					}
				}
			}
		}
	}
	return out
}

// receiverName is the name of an exported receiver type, or "" for a method
// of an unexported type, which callers cannot name.
func receiverName(recv *ast.FieldList) string {
	if len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if idx, ok := expr.(*ast.IndexExpr); ok { // a generic type
		expr = idx.X
	}
	if idx, ok := expr.(*ast.IndexListExpr); ok {
		expr = idx.X
	}
	if id, ok := expr.(*ast.Ident); ok && id.IsExported() {
		return id.Name
	}
	return ""
}

// memberNames are the exported fields of a struct and the exported methods of
// an interface, as "Type.Member".
func memberNames(typeName string, expr ast.Expr) []string {
	var fields *ast.FieldList
	switch t := expr.(type) {
	case *ast.StructType:
		fields = t.Fields
	case *ast.InterfaceType:
		fields = t.Methods
	}
	if fields == nil {
		return nil
	}
	var out []string
	for _, f := range fields.List {
		for _, n := range f.Names {
			if n.IsExported() {
				out = append(out, typeName+"."+n.Name)
			}
		}
	}
	return out
}
