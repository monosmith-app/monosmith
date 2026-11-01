package sources

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/tui/adoptwizard"
)

// AdoptionOpener reads the repository in dir for the adopt wizard. It runs
// off the UI goroutine: it looks at git, the project files and, for an Nx
// workspace, asks Nx to describe the projects. It says how many files it is
// about to read, so the wait has a size.
func AdoptionOpener(dir, version string) adoptwizard.Opener {
	return func(out io.Writer) (adoptwizard.Flow, error) {
		if n, err := adoptflow.CountFiles(dir); err == nil {
			fmt.Fprintln(out, adoptflow.ReadingLine(n))
		}
		// Each stage the read reports becomes the step the dialog shows.
		s, err := adoptflow.OpenWithProgress(dir, version, func(stage string) {
			fmt.Fprintln(out, "… "+strings.TrimSuffix(stage, "…"))
		})
		if err != nil {
			return nil, err
		}
		return s, nil
	}
}
