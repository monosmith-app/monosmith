package depsrepair

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// kept are the names at the workspace root a repair never deletes, whatever a
// plugin says: the history and the workspace's own configuration.
var kept = map[string]bool{".git": true, ".smith": true, "smith.toml": true}

// checkDeletePath says whether smith deletes a path a repair plan names, and
// where it is on disk. It refuses:
//   - a path that is not a clean workspace-relative path below the root (empty,
//     ".", absolute, with a backslash, a drive or a ".." step);
//   - .git, .smith and smith.toml at the workspace root;
//   - a project's root, or a folder that holds a project;
//   - a path below a link (a symlink or a Windows junction): the deletion would
//     go through the link and leave the workspace.
//
// The path itself may be a link: it is then deleted as a link.
//
// Parameters: root is the absolute workspace root; rel is the path as the plan
// gives it; projectRoots are the workspace-relative roots of the projects.
//
// Returns: the absolute path to delete.
//
// Errors: one sentence that names the path and the rule it breaks.
func checkDeletePath(root, rel string, projectRoots []string) (string, error) {
	if !insideWorkspace(rel) {
		return "", fmt.Errorf("%q is not a path inside the workspace (a workspace-relative path with forward slashes, without ..)", rel)
	}
	parts := strings.Split(rel, "/")
	if kept[parts[0]] {
		return "", fmt.Errorf("%s belongs to the workspace itself, not to a tool: a repair never deletes it", rel)
	}
	for _, project := range projectRoots {
		project = path.Clean(filepath.ToSlash(project))
		if project == "." || project == "" {
			continue
		}
		if project == rel || strings.HasPrefix(project, rel+"/") {
			return "", fmt.Errorf("%s holds the project in %s: a repair deletes installed dependencies, never source", rel, project)
		}
	}
	at := root
	for _, part := range parts[:len(parts)-1] {
		at = filepath.Join(at, part)
		info, err := os.Lstat(at)
		if err != nil {
			// Nothing is there, so nothing below it is deleted either.
			break
		}
		if isLink(info.Mode()) {
			return "", fmt.Errorf("%s is below %s, which is a link: a repair never deletes through a link", rel, filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(at, root), string(filepath.Separator))))
		}
	}
	return filepath.Join(root, filepath.FromSlash(rel)), nil
}

// insideWorkspace reports whether rel is a clean workspace-relative path below
// the root: the rule the plugin conformance check applies to a plan.
func insideWorkspace(rel string) bool {
	if rel == "" || strings.ContainsAny(rel, "\\:") || strings.HasPrefix(rel, "/") {
		return false
	}
	clean := path.Clean(rel)
	return clean == rel && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}

// isLink reports whether a mode read without following is a link: a symlink,
// or a Windows reparse point that is not one (a junction, which Go reports as
// irregular). Neither is a folder smith walks into.
func isLink(mode fs.FileMode) bool {
	return mode&(fs.ModeSymlink|fs.ModeIrregular) != 0
}
