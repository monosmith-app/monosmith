package scaffold

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// UnsavedFiles are the files under dir that git cannot bring back: untracked
// files (not ignored) and tracked files with uncommitted changes. Deleting
// the folder loses them for good. Ignored files (node_modules, build output)
// are counted apart, because a build or an install makes them again.
type UnsavedFiles struct {
	// Files are paths relative to the workspace root, sorted by git.
	Files []string
	// Ignored is how many ignored entries the folder holds.
	Ignored int
}

// ListUnsavedFiles asks git what under dir (a path relative to the workspace
// root) would be lost for good if the folder were deleted. It returns nothing
// when root is not a git repository, because then smith cannot tell.
//
// Parameters: root is the workspace root; dir is the folder, slash-separated.
func ListUnsavedFiles(root, dir string) UnsavedFiles {
	git := gitclient.New(root)
	if !git.IsRepository() {
		return UnsavedFiles{}
	}
	out, err := git.Run("status", "--porcelain", "--untracked-files=all", "--ignored", "--", dir)
	if err != nil {
		return UnsavedFiles{}
	}
	return parseUnsaved(out)
}

// parseUnsaved reads `git status --porcelain` lines: "!!" is ignored, anything
// else with a path is unsaved work.
func parseUnsaved(porcelain string) UnsavedFiles {
	var u UnsavedFiles
	for _, line := range strings.Split(porcelain, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.Trim(line[3:], `"`)
		if line[:2] == "!!" {
			u.Ignored++
			continue
		}
		u.Files = append(u.Files, path)
	}
	return u
}
