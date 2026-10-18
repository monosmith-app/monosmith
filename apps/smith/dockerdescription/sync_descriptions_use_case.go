package dockerdescription

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-docker/publishing"
)

// Request says which apps to sync and how a failure counts.
type Request struct {
	// Projects limits the sync to these apps; empty means every app.
	Projects []string
	DryRun   bool
	// Strict makes a failed description a failed command; otherwise it is
	// only a warning, as in the release.
	Strict bool
}

// ErrFailed is returned, with Strict, when a description could not be set.
var ErrFailed = errors.New("a Docker Hub description could not be set")

// syncDescription is the docker plugin's description sync; tests replace it.
var syncDescription = publishing.SyncDescription

// Run sets the description of each app that names a description_from file,
// printing what it did. An app with no docker.io registry or no file is left
// out of the output, unless it was asked for by name.
func Run(root string, projects []*projectgraph.Project, req Request, out io.Writer) error {
	var failed []string
	attempted := 0
	for _, p := range projects {
		asked := slices.Contains(req.Projects, p.Name)
		if len(req.Projects) > 0 && !asked {
			continue
		}
		res, err := syncDescription(root, p.Project, req.DryRun)
		if err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		if res.Skipped != "" {
			if asked {
				fmt.Fprintf(out, "%s: nothing to set (%s)\n", p.Name, res.Skipped)
			}
			continue
		}
		attempted++
		verb := "set the Docker Hub description of"
		if req.DryRun {
			verb = "would set the Docker Hub description of"
		}
		for _, repository := range res.Set {
			fmt.Fprintf(out, "%s: %s %s\n", p.Name, verb, repository)
		}
		for _, f := range res.Failed {
			fmt.Fprintf(out, "%s: could not set the Docker Hub description: %s\n", p.Name, f)
			failed = append(failed, p.Name+": "+f)
		}
	}
	if attempted == 0 && len(req.Projects) == 0 {
		fmt.Fprintln(out, "no app names a description_from file for a docker.io registry")
	}
	if len(failed) > 0 && req.Strict {
		return fmt.Errorf("%w: %s", ErrFailed, strings.Join(failed, "; "))
	}
	return nil
}
