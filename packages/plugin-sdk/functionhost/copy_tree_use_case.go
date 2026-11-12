package functionhost

import (
	"io/fs"
	"os"
	"path/filepath"
)

// copyTree copies a file or a folder with everything in it. A folder named
// node_modules is left out.
func copyTree(from, to string) error {
	return copyTreeSkipping(from, to, func(name string, isDir bool) bool { return isDir && name == "node_modules" })
}

// copyTreeSkipping is copyTree with the entries that skip names left out: a
// folder it skips is not entered.
func copyTreeSkipping(from, to string, skip func(name string, isDir bool) bool) error {
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		if p != from && skip(d.Name(), d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		return copyFile(p, filepath.Join(to, rel))
	})
}
