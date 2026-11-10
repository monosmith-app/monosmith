package releaseassets

import (
	"context"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// runFromTag builds and uploads the files of one tag from the tag's own
// commit. It checks the commit out in a temporary git worktree, opens that as a
// workspace, runs Prepare there (the install), runs the project's build and
// removes the worktree. The worktree is removed on every path, including a
// failed build.
//
// It returns a problem when the commit has no smith.toml, or its smith.toml
// does not give the project a release build: the config did not exist yet, so
// only --source current can build it.
func runFromTag(ctx context.Context, s *session.Session, req Request, project, platform, tag, commit string) error {
	short := commit
	if len(short) > 10 {
		short = short[:10]
	}
	noConfig := problem.New(fmt.Sprintf("the commit %s of %s has no release build for %s: its smith.toml did not have the assets config yet", short, tag, project),
		"run it again with --source current, which builds the code as it is now with the version of the tag")
	if !hasConfigAt(s.Git, commit) {
		return noConfig
	}
	return withTagWorktree(s.Git, tag, func(dir string) error {
		fmt.Fprintf(req.Out, "building from the commit %s of %s in %s\n", short, tag, dir)
		ws, err := session.Open(session.Options{Dir: dir, Exact: true, Version: s.Version, Log: req.Out})
		if err != nil {
			return problem.New(fmt.Sprintf("the smith.toml of %s cannot be read: %v", tag, err), "run it again with --source current")
		}
		defer ws.Close()
		found := false
		for _, p := range ws.Graph.Projects {
			if p.Name == project && release.BuildsAssetsFromCommand(p) {
				found = true
			}
		}
		if !found {
			return noConfig
		}
		if req.Prepare != nil {
			if err := req.Prepare(ctx, ws); err != nil {
				return err
			}
		}
		_, err = Run(ctx, ws, Request{Platform: platform, Project: project, Tag: tag, Source: SourceCurrent, Out: req.Out})
		return err
	})
}
