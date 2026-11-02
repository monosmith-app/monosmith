package dependencies

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// maxSnapshotBytes is the largest file the guard keeps a copy of. A manifest
// or a lock file is far smaller; a bigger file is a build output.
const maxSnapshotBytes = 4 << 20

// guarded runs change, which adds packages and so edits the manifest and lock
// files in dirs (the project folder and the workspace root). When change
// fails, the files that changed come back to what they held before, so a
// package name that does not exist cannot stay in the manifest and break every
// later install. The returned error is the failure of change, and it says which
// files were put back.
func guarded(dirs []string, change func() error) error {
	before := snapshot(dirs)
	err := change()
	if err == nil {
		return nil
	}
	if restored := before.restore(); len(restored) > 0 {
		return fmt.Errorf("%w\nNothing was kept: smith put %s back as they were. You can:\n  1. Try again with the exact package name of the public registry (the default).\n  2. Cancel: your project is as it was before.\n  A library of this workspace is not on the registry, so its name cannot be installed this way", err, strings.Join(restored, ", "))
	}
	return err
}

// snapshotFiles is the content of the files directly in some folders.
type snapshotFiles map[string][]byte

// snapshot copies the regular files directly inside each dir; a folder that
// cannot be read, or a file that is too big, is left out.
func snapshot(dirs []string) snapshotFiles {
	out := snapshotFiles{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			info, err := e.Info()
			if err != nil || !info.Mode().IsRegular() || info.Size() > maxSnapshotBytes {
				continue
			}
			path := filepath.Join(dir, e.Name())
			if data, err := os.ReadFile(path); err == nil {
				out[path] = data
			}
		}
	}
	return out
}

// restore writes back each file whose content changed or that was deleted, and
// returns the names of the files it put back, sorted.
func (s snapshotFiles) restore() []string {
	var names []string
	for path, data := range s {
		if now, err := os.ReadFile(path); err == nil && bytes.Equal(now, data) {
			continue
		}
		if os.WriteFile(path, data, 0o644) == nil {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names
}
