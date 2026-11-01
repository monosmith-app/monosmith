package convert

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// snapshot remembers files before a conversion changes them, so a conversion
// that fails half way can put them back. It keeps the content of each file it
// is asked about, and which ones did not exist.
type snapshot struct {
	root   string
	saved  map[string]savedFile
	absent map[string]bool
}

func newSnapshot(root string) *snapshot {
	return &snapshot{root: root, saved: map[string]savedFile{}, absent: map[string]bool{}}
}

// skipped are folders the snapshot never reads: caches and installed packages
// are not part of the repository, and the conversion does not need them back.
func skipped(name string) bool { return name == "node_modules" || name == ".nx" || name == ".git" }

// remember keeps the file at rel (workspace-relative) or, for a folder, every
// file in it. A path that does not exist is remembered as absent.
func (s *snapshot) remember(rel string) {
	full := filepath.Join(s.root, filepath.FromSlash(rel))
	info, err := os.Stat(full)
	if errors.Is(err, fs.ErrNotExist) {
		s.absent[rel] = true
		return
	}
	if err != nil {
		return
	}
	if !info.IsDir() {
		if _, done := s.saved[rel]; done {
			return
		}
		if raw, err := os.ReadFile(full); err == nil {
			s.saved[rel] = savedFile{raw: raw, mode: info.Mode().Perm()}
		}
		return
	}
	_ = filepath.WalkDir(full, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipped(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if r, err := filepath.Rel(s.root, path); err == nil {
			s.remember(filepath.ToSlash(r))
		}
		return nil
	})
}

// restore writes every remembered file back and removes the files that did not
// exist before. It returns the first error and keeps going.
func (s *snapshot) restore() error {
	var first error
	keep := func(err error) {
		if err != nil && first == nil {
			first = err
		}
	}
	for rel, f := range s.saved {
		full := filepath.Join(s.root, filepath.FromSlash(rel))
		keep(os.MkdirAll(filepath.Dir(full), 0o755))
		keep(os.WriteFile(full, f.raw, f.mode))
	}
	for rel := range s.absent {
		keep(os.RemoveAll(filepath.Join(s.root, filepath.FromSlash(rel))))
	}
	return first
}

// savedFile is one remembered file.
type savedFile struct {
	raw  []byte
	mode fs.FileMode
}
