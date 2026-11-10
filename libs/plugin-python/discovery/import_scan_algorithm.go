package discovery

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	skippedDirs   = map[string]bool{"dist": true, "build": true, "node_modules": true, "__pycache__": true}
	importPattern = regexp.MustCompile(`(?m)^[ \t]*(?:from[ \t]+([A-Za-z_][\w.]*)[ \t]+import\b|import[ \t]+([A-Za-z_][\w.]*(?:[ \t]*,[ \t]*[A-Za-z_][\w.]*)*))`)
)

// ScanImports returns the top-level module names imported by the .py files
// under dir, sorted and unique. Relative imports are ignored. Sdist staging
// folders are skipped, and a file or folder that vanishes during the walk (a
// build removing its staging folder) is ignored.
func ScanImports(dir string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && path != dir {
				return nil
			}
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != dir && (skippedDirs[name] || strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".egg-info") || IsSdistStaging(name)) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".py" {
			return nil
		}
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		for _, name := range ModuleNames(string(src)) {
			seen[name] = true
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, err
}

// ModuleNames extracts top-level module names from one file's imports.
func ModuleNames(src string) []string {
	var out []string
	for _, m := range importPattern.FindAllStringSubmatch(src, -1) {
		modules := m[1]
		if modules == "" {
			modules = m[2]
		}
		for _, mod := range strings.Split(modules, ",") {
			top, _, _ := strings.Cut(strings.TrimSpace(mod), ".")
			if top != "" {
				out = append(out, top)
			}
		}
	}
	return out
}
