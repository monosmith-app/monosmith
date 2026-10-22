package components

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// File is a lockfile or manifest read from disk.
type File struct {
	// Rel is the workspace-relative path, forward slashes.
	Rel  string
	Body []byte
}

// FindUp looks for the first of names in the project's directory, then in
// each directory above it up to the workspace root (lockfiles of a workspace
// sit at its root). Names are tried in order within a directory. ok is false
// when none exists.
func FindUp(root, projectRoot string, names ...string) (File, bool, error) {
	rel := path.Clean(strings.TrimPrefix(filepath.ToSlash(projectRoot), "./"))
	for {
		for _, name := range names {
			p := path.Join(rel, name)
			body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
			if err == nil {
				return File{Rel: p, Body: body}, true, nil
			}
			if !os.IsNotExist(err) {
				return File{}, false, err
			}
		}
		if rel == "." || rel == "/" || rel == "" {
			return File{}, false, nil
		}
		rel = path.Dir(rel)
	}
}

// ReadIn reads one file of the project's own directory; ok is false when it
// does not exist.
func ReadIn(root, projectRoot, name string) (File, bool, error) {
	p := path.Join(path.Clean(filepath.ToSlash(projectRoot)), name)
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
	if os.IsNotExist(err) {
		return File{}, false, nil
	}
	if err != nil {
		return File{}, false, err
	}
	return File{Rel: p, Body: body}, true, nil
}
