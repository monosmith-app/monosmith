package declaredimports

import (
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Check returns the imports of the Go module at project (a path relative to
// the workspace root) that no module its go.mod requires provides, sorted by
// file then package. See doc.go for what is never reported.
func Check(workspaceRoot, project string) ([]UndeclaredImport, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(project))
	src, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, err
	}
	mod := parseModuleFile(string(src))
	if mod.module == "" {
		return nil, errors.New(project + "/go.mod has no module line")
	}
	provided := append(append([]string{mod.module}, mod.required...), mod.replaced...)
	provided = append(provided, workspaceModules(workspaceRoot)...)

	var out []UndeclaredImport
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if p == dir {
				return nil
			}
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
				return filepath.SkipDir // a nested module has its own go.mod
			}
			return nil
		}
		if filepath.Ext(name) != ".go" {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly|parser.ParseComments)
		if err != nil || ignored(file.Comments) {
			return nil // a file that does not parse is the compiler's to report
		}
		rel, _ := filepath.Rel(dir, p)
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil || isStandard(path) || providedBy(provided, path) {
				continue
			}
			out = append(out, UndeclaredImport{File: filepath.ToSlash(rel), Package: path})
		}
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Package < out[j].Package
	})
	return out, err
}

// workspaceModules are the module paths of the go.work's use directives: the
// workspace provides them to every module in it.
func workspaceModules(workspaceRoot string) []string {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	if err != nil {
		return nil
	}
	work := parseModuleFile(string(src))
	modules := work.replaced
	for _, use := range work.used {
		if mod, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(use), "go.mod")); err == nil {
			modules = append(modules, parseModuleFile(string(mod)).module)
		}
	}
	return modules
}

// isStandard reports whether an import path is the standard library's (or cgo's
// "C"): its first element has no dot.
func isStandard(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

func providedBy(modules []string, path string) bool {
	for _, m := range modules {
		if path == m || strings.HasPrefix(path, m+"/") {
			return true
		}
	}
	return false
}
