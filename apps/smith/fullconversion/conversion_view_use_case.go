package fullconversion

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

// View is the workspace as the conversion would leave it, made on a scratch
// copy so the real files stay as they are until the user agrees.
type View struct {
	Tree *legacyinventory.Tree
	// Dir is the scratch copy; it is removed by Close.
	Dir string
}

// Close removes the scratch copy.
func (v *View) Close() { _ = os.RemoveAll(v.Dir) }

// MakeView copies the workspace's files to a scratch folder and has
// materialize apply the conversion there. The tree it returns reads the
// scratch copy and describes installed packages from the real workspace's
// node_modules.
func MakeView(root string, files []string, materialize func(dir string) error) (*View, error) {
	dir, err := os.MkdirTemp("", "smith-convert-view-")
	if err != nil {
		return nil, err
	}
	view := &View{Dir: dir}
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil {
			continue // a file that disappeared is no longer part of the workspace
		}
		target := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			view.Close()
			return nil, err
		}
		if err := os.WriteFile(target, raw, 0o644); err != nil {
			view.Close()
			return nil, err
		}
	}
	// The conversion makes its changes through git; an empty repository is
	// enough for it to work in.
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		view.Close()
		return nil, err
	}
	if err := materialize(dir); err != nil {
		view.Close()
		return nil, err
	}
	var converted []string
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() && (rel == ".git" || rel == "node_modules" || strings.HasSuffix(rel, "/node_modules")) {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			converted = append(converted, rel)
		}
		return nil
	})
	if err != nil {
		view.Close()
		return nil, err
	}
	view.Tree = legacyinventory.NewTree(dir, converted)
	view.Tree.Modules = root
	return view, nil
}
