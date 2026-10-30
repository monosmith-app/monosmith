package declaredimports

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// sourceImport is one import of one file.
type sourceImport struct {
	File string
	Path string
}

var generateTool = regexp.MustCompile(`(?m)^//go:generate\s+go\s+(?:run|tool)\s+(?:-\S+\s+)*([^\s@]+)`)

// skippedDir is a directory the go command does not look into.
func skippedDir(name string) bool {
	return name == "testdata" || name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanImports lists the imports of every .go file under dir, test files and
// files for any build tag included (go mod tidy looks at all of them), except
// files tagged "ignore", testdata, vendor, and nested modules. A go:generate
// line that runs a package (go run example.com/tool) counts as importing it.
// A file that does not parse is skipped: the compiler reports it.
func scanImports(dir string) ([]sourceImport, error) {
	var out []sourceImport
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == dir {
				return nil
			}
			if skippedDir(d.Name()) {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.ImportsOnly|parser.ParseComments)
		if err != nil || ignored(file) {
			return nil
		}
		for _, spec := range file.Imports {
			if p, err := strconv.Unquote(spec.Path.Value); err == nil {
				out = append(out, sourceImport{File: rel, Path: p})
			}
		}
		for _, m := range generateTool.FindAllStringSubmatch(string(src), -1) {
			out = append(out, sourceImport{File: rel, Path: m[1]})
		}
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Path < out[j].Path
	})
	return out, err
}

// ignored reports a file the go command leaves out of every build: one whose
// constraint before the package clause is exactly "ignore".
func ignored(file *ast.File) bool {
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			break
		}
		for _, c := range group.List {
			if constraint, ok := strings.CutPrefix(c.Text, "//go:build"); ok && strings.TrimSpace(constraint) == "ignore" {
				return true
			}
		}
	}
	return false
}
