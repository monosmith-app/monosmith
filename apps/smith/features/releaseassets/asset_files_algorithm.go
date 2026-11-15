package releaseassets

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// assetFiles expands workspace-relative globs against fsys into absolute
// paths under root, sorted and without repeats. A glob may match nothing, as
// a .dmg glob does on Windows, when another glob of the list matched; the
// globs that matched nothing are returned as skipped. When no glob matched,
// the build left nothing of what the release promised, and that is an error
// that names the first glob.
func assetFiles(fsys fs.FS, root string, globs []string) (files, skipped []string, err error) {
	for _, g := range globs {
		matches, err := doublestar.Glob(fsys, g, doublestar.WithFilesOnly())
		if err != nil {
			return nil, nil, fmt.Errorf("assets %q: %w", g, err)
		}
		if len(matches) == 0 {
			skipped = append(skipped, g)
			continue
		}
		for _, m := range matches {
			files = append(files, filepath.Join(root, filepath.FromSlash(m)))
		}
	}
	if len(files) == 0 && len(skipped) > 0 {
		return nil, nil, fmt.Errorf("the build left no file matching assets \"%s\"", strings.Join(skipped, `", "`))
	}
	slices.Sort(files)
	return slices.Compact(files), skipped, nil
}
