package sources

import (
	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
)

// AdoptionOpener reads the repository in dir for the adopt wizard. It runs
// off the UI goroutine: it looks at git, the project files and, for an Nx
// workspace, asks Nx to describe the projects.
func AdoptionOpener(dir, version string) adoptwizard.Opener {
	return func() (adoptwizard.Flow, error) {
		s, err := adoptflow.Open(dir, version)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
}
