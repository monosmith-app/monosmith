package outsidefiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Bring copies each movable file into the project folder under its base
// name, then writes the new path in the manifest that names it: every
// occurrence of the old path (..\LICENSE) becomes the name (LICENSE). A file
// whose name the folder holds already, or whose manifest cannot be written,
// does not come along.
//
// Parameters: dir is the project folder on disk; files are the files Find
// returned (only the movable ones are copied).
// Returns: the files that came along, and those that did not, each with the
// reason.
func Bring(dir string, files []File) (brought []File, failed map[string]error) {
	failed = map[string]error{}
	for _, f := range files {
		if !f.Movable() {
			continue
		}
		if err := bring(dir, f); err != nil {
			failed[f.Path] = err
			continue
		}
		brought = append(brought, f)
	}
	return brought, failed
}

func bring(dir string, f File) error {
	target := filepath.Join(dir, f.Name())
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("the project folder has a file %s already", f.Name())
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	manifest := filepath.Join(dir, f.Manifest)
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	if err := os.WriteFile(target, f.Content, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(manifest, []byte(strings.ReplaceAll(string(raw), f.Ref, f.Name())), 0o644); err != nil {
		_ = os.Remove(target)
		return err
	}
	return nil
}
