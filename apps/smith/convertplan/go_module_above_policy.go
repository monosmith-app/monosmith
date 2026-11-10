package convertplan

import (
	"path"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// GoModuleAbove says whether an Nx project with no manifest is a Go package of
// a module declared above it: it holds .go files and a go.mod sits in a parent
// folder (files are the workspace's, workspace-relative). dir is that module's
// folder, "the repository root" for the root. smith runs a Go project from a
// go.mod in its own folder, so such a project has no plugin until it has one.
func GoModuleAbove(p *nxworkspace.Project, files []string) (dir string, ok bool) {
	holdsGo := false
	for _, f := range p.Files {
		if strings.HasPrefix(f, p.Root+"/") && strings.HasSuffix(f, ".go") {
			holdsGo = true
			break
		}
	}
	if !holdsGo {
		return "", false
	}
	for parent := path.Dir(p.Root); ; parent = path.Dir(parent) {
		manifest := "go.mod"
		if parent != "." {
			manifest = parent + "/go.mod"
		}
		if slices.Contains(files, manifest) {
			if parent == "." {
				return "the repository root", true
			}
			return parent, true
		}
		if parent == "." || parent == "/" {
			return "", false
		}
	}
}
