package pathfinder

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// maxNames is how many names of one folder are read. A folder with more is
// shown as far as read and says so: the list never waits for a huge folder.
const maxNames = 5000

// rawEntry is one name found in a folder.
type rawEntry struct {
	name string
	dir  bool
}

// listFolder reads the names of the folder dir (slash-separated, relative to
// root) and no deeper.
//
// Parameters: root is the folder smith started in; dir is the folder to read.
// Returns: the names (at most maxNames), whether there were more, and the error
// when the folder cannot be read.
func listFolder(root, dir string) (names []rawEntry, truncated bool, err error) {
	path := filepath.Join(root, filepath.FromSlash(dir))
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = f.Close() }()
	entries, err := f.ReadDir(maxNames + 1)
	if err != nil && !errors.Is(err, io.EOF) && len(entries) == 0 {
		return nil, false, err
	}
	for _, e := range entries {
		isDir := e.IsDir()
		if e.Type()&os.ModeSymlink != 0 {
			if st, err := os.Stat(filepath.Join(path, e.Name())); err == nil {
				isDir = st.IsDir()
			}
		}
		names = append(names, rawEntry{name: e.Name(), dir: isDir})
	}
	if len(names) > maxNames {
		return names[:maxNames], true, nil
	}
	return names, false, nil
}
