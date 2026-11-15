package pathfinder

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// buildFolders are the folders that hold what a build or an install makes.
// They are dimmed even when no .gitignore names them.
var buildFolders = []string{"node_modules", "dist", "build", "target", "coverage", "__pycache__", ".venv", "venv"}

// ignoreRules says which names git ignores, as far as the .gitignore of the
// folder smith started in tells (plain names, globs and folder names; it does
// not run git).
type ignoreRules struct {
	names    []string
	anchored []string
}

// loadIgnoreRules reads the .gitignore at the root; a missing file gives the
// build folders only.
func loadIgnoreRules(root string) ignoreRules {
	r := ignoreRules{names: append([]string{}, buildFolders...)}
	f, err := os.Open(filepath.Join(root, ".gitignore"))
	if err != nil {
		return r
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		switch {
		case line == "":
		case strings.Contains(line, "/"):
			r.anchored = append(r.anchored, strings.TrimPrefix(line, "/"))
		default:
			r.names = append(r.names, line)
		}
	}
	return r
}

// Ignored reports whether the path rel (slash-separated, relative to the root)
// is ignored: the name itself, or any folder above it, matches a rule.
func (r ignoreRules) Ignored(rel string) bool {
	parts := strings.Split(strings.TrimSuffix(rel, "/"), "/")
	for i := range parts {
		if parts[i] == "" || parts[i] == ".." || parts[i] == "." {
			continue
		}
		for _, n := range r.names {
			if ok, _ := path.Match(n, parts[i]); ok {
				return true
			}
		}
		for _, a := range r.anchored {
			if ok, _ := path.Match(a, strings.Join(parts[:i+1], "/")); ok {
				return true
			}
		}
	}
	return false
}
