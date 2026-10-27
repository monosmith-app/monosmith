package fileedits

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Set is the edits made so far to the files of one workspace. Nothing is
// written to disk.
type Set struct {
	root  string
	files map[string]*entry
}

type entry struct {
	original string
	current  string
}

// New starts an empty set over the workspace at root (an OS path).
func New(root string) *Set {
	return &Set{root: root, files: map[string]*entry{}}
}

// Read returns a workspace-relative file's content with the edits so far.
// ok is false when the file does not exist.
func (s *Set) Read(rel string) (content string, ok bool) {
	if e, found := s.files[rel]; found {
		return e.current, true
	}
	raw, err := os.ReadFile(filepath.Join(s.root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	text := string(raw)
	s.files[rel] = &entry{original: text, current: text}
	return text, true
}

// Write sets a file's content. The file must exist: edits change files.
func (s *Set) Write(rel, content string) {
	if _, ok := s.Read(rel); !ok {
		return
	}
	s.files[rel].current = content
}

// Edits are the files whose content differs from what is on disk, by path.
func (s *Set) Edits() []protocol.FileEdit {
	var out []protocol.FileEdit
	for rel, e := range s.files {
		if e.current != e.original {
			out = append(out, protocol.FileEdit{Path: rel, Content: e.current})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// List lists the workspace-relative files directly in dir ("" is the
// workspace root) whose names pass keep; it does not look into folders.
func (s *Set) List(dir string, keep func(name string) bool) []string {
	entries, err := os.ReadDir(filepath.Join(s.root, filepath.FromSlash(dir)))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && keep(e.Name()) {
			out = append(out, path.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out
}

// Walk lists the workspace-relative files under dir (forward slashes) whose
// names pass keep, skipping dependency and build directories.
func (s *Set) Walk(dir string, keep func(name string) bool) []string {
	var out []string
	base := filepath.Join(s.root, filepath.FromSlash(dir))
	_ = filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".git", "target", "vendor", "dist", "build", ".venv", "__pycache__":
				return filepath.SkipDir
			}
			return nil
		}
		if keep(d.Name()) {
			rel, err := filepath.Rel(s.root, p)
			if err == nil {
				out = append(out, strings.ReplaceAll(rel, string(filepath.Separator), "/"))
			}
		}
		return nil
	})
	sort.Strings(out)
	return out
}
