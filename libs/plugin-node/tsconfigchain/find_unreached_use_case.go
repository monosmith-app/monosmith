package tsconfigchain

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Unreached lists the tsconfig*.json files of a workspace whose extends chain
// does not reach the root tsconfig.base.json, as paths relative to root with
// slashes. The root files of the chain, folders that hold no source of the
// workspace (node_modules, dist, folders that start with a dot) and files that
// are not JSON are skipped.
func Unreached(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (name == "node_modules" || name == "dist" || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if !strings.HasPrefix(name, "tsconfig") || !strings.HasSuffix(name, ".json") || filepath.Dir(rel) == "." {
			return nil
		}
		if ok, err := ReachesBase(root, path); err == nil && !ok {
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out, err
}
