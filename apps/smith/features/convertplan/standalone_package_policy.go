package convertplan

import (
	"path"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/monosmith-app/monosmith/apps/smith/features/nxworkspace"
)

// Standalone reports whether an Nx project is a Node package the workspace
// deliberately keeps apart: outside the root package.json "workspaces", so
// npm never installs it with the workspace, and known to Nx only through its
// package.json (no project.json, not released). smith leaves such a
// directory out of the workspace instead of making it a project npm cannot
// install.
func Standalone(w *nxworkspace.Workspace, p *nxworkspace.Project, released bool) bool {
	if released || p.HasProjectJSON || p.Package == nil || p.Root == "." || p.Root == "" {
		return false
	}
	if lang, _ := Language(p); lang != "node" {
		return false
	}
	globs := w.RootPackage.WorkspaceGlobs()
	if len(globs) == 0 {
		return false
	}
	for _, g := range globs {
		g = strings.TrimSuffix(strings.TrimPrefix(g, "./"), "/")
		if ok, _ := doublestar.Match(g, p.Root); ok {
			return false
		}
	}
	return true
}

// ExcludeGlobs are the [workspace] exclude entries for the standalone roots:
// "<parent>/*" when every project under that parent is standalone (the
// examples/ of a repository), the root itself otherwise.
func ExcludeGlobs(standalone, kept []string) []string {
	byParent := map[string][]string{}
	for _, root := range standalone {
		byParent[path.Dir(root)] = append(byParent[path.Dir(root)], root)
	}
	seen := map[string]bool{}
	var globs []string
	for parent, roots := range byParent {
		shared := parent == "."
		for _, k := range kept {
			if strings.HasPrefix(k, parent+"/") {
				shared = true
			}
		}
		if !shared {
			globs = append(globs, parent+"/*")
			continue
		}
		for _, r := range roots {
			if !seen[r] {
				seen[r] = true
				globs = append(globs, r)
			}
		}
	}
	sort.Strings(globs)
	return globs
}
