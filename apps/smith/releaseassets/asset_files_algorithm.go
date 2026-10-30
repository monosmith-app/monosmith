package releaseassets

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
)

// assetFiles expands workspace-relative globs against fsys into absolute
// paths under root, sorted and without repeats. A glob that matches nothing
// is an error: the build did not leave what the release promised.
func assetFiles(fsys fs.FS, root string, globs []string) ([]string, error) {
	var files []string
	for _, g := range globs {
		matches, err := doublestar.Glob(fsys, g, doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("assets %q: %w", g, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("the build left no file matching assets %q", g)
		}
		for _, m := range matches {
			files = append(files, filepath.Join(root, filepath.FromSlash(m)))
		}
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}
