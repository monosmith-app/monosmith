package turboworkspace

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// NxShape describes the packages the way the Nx conversion expects its
// projects, so the shared parts of the conversion read both. A package whose
// directory says it is an app (apps/) is an application.
func (w *Workspace) NxShape() *nxworkspace.Workspace {
	n := &nxworkspace.Workspace{Root: w.Root, NxVersion: "", RootPackage: w.RootPackage}
	if len(w.Globs) > 0 {
		n.RootPackage.Workspaces, _ = json.Marshal(positive(w.Globs))
	}
	for _, p := range w.Packages {
		pkg := p.JSON
		np := &nxworkspace.Project{
			Name:    p.Name,
			Root:    p.Dir,
			Type:    "lib",
			Targets: map[string]nxworkspace.Target{},
			Files:   p.Files,
			Package: &pkg,
		}
		if top, _, _ := strings.Cut(p.Dir, "/"); top == "apps" {
			np.Type, np.ProjectType = "app", "application"
		}
		n.Projects = append(n.Projects, np)
	}
	return n
}

func positive(globs []string) []string {
	var out []string
	for _, g := range globs {
		if !strings.HasPrefix(g, "!") {
			out = append(out, g)
		}
	}
	return out
}
