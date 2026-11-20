package workspacejoin

import (
	"path"
	"strings"
)

// Covers says whether one of the "workspaces" globs matches the project
// folder root, as npm matches them: a glob such as "apps/*", or the folder
// itself, with or without a leading "./".
//
// Parameters: globs are the root package.json "workspaces"; root is the
// project's workspace-relative folder.
// Returns: true when npm install reaches the project.
func Covers(globs []string, root string) bool {
	for _, g := range globs {
		g = strings.TrimSuffix(strings.TrimPrefix(g, "./"), "/")
		if g == root {
			return true
		}
		if ok, _ := path.Match(g, root); ok {
			return true
		}
	}
	return false
}

// Pattern is the "workspaces" entry that joins the project at root: the glob
// of its parent folder ("apps/*"), the form smith's root package.json uses, or
// the folder itself for a project at the top of the workspace.
//
// Parameters: root is the project's workspace-relative folder.
// Returns: the entry to add.
func Pattern(root string) string {
	if dir := path.Dir(root); dir != "." {
		return dir + "/*"
	}
	return root
}
