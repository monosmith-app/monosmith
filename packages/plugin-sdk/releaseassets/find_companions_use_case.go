package releaseassets

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// companionPrefixes are the files an archive carries beside the program: the
// README and the licence, under the names projects give them.
var companionPrefixes = []string{"README", "LICENSE", "LICENCE", "COPYING"}

// Companions finds the README and licence files to put in an archive: the
// ones in the project's directory, else the ones at the workspace root, each
// kind from one place only. The result is sorted, so archives are the same
// every time.
func Companions(workspaceRoot, projectRoot string) []Entry {
	var found []Entry
	for _, prefix := range companionPrefixes {
		for _, dir := range []string{filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot)), workspaceRoot} {
			files := withPrefix(dir, prefix)
			if len(files) == 0 {
				continue
			}
			for _, f := range files {
				found = append(found, Entry{Path: filepath.Join(dir, f), Name: f})
			}
			break
		}
	}
	slices.SortFunc(found, func(a, b Entry) int { return strings.Compare(a.Name, b.Name) })
	return found
}

func withPrefix(dir, prefix string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(strings.ToUpper(e.Name()), prefix) {
			continue
		}
		names = append(names, e.Name())
	}
	return names
}
