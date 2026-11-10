package discovery

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

var skippedDirs = map[string]bool{"vendor": true, "testdata": true, "dist": true, "node_modules": true}

// ScanImports returns every import path used by the Go files under dir,
// sorted and unique. Nested modules (another go.mod) are not descended into.
func ScanImports(dir string) ([]string, error) {
	seen := map[string]bool{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && path != dir {
				return nil // removed while the walk ran
			}
			return err
		}
		if d.IsDir() {
			if path == dir {
				return nil
			}
			if skippedDirs[d.Name()] || strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") {
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
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil // a file that does not parse is go build's problem, not the graph's
		}
		for _, spec := range file.Imports {
			if p, err := strconv.Unquote(spec.Path.Value); err == nil {
				seen[p] = true
			}
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, err
}
