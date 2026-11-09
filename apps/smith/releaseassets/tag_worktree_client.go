package releaseassets

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// hasConfigAt reports whether the commit holds a smith.toml.
func hasConfigAt(g *gitclient.Git, commit string) bool {
	_, err := g.Run("cat-file", "-e", commit+":smith.toml")
	return err == nil
}

// withTagWorktree checks the tag out in a temporary detached git worktree, runs
// fn with its directory and removes the worktree afterwards, whatever fn
// returns. The checkout the caller works in is not touched.
//
// It returns a problem when git cannot make the worktree, or the error of fn.
func withTagWorktree(g *gitclient.Git, tag string, fn func(dir string) error) error {
	dir, err := os.MkdirTemp("", "smith-assets-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if _, err := g.Run("worktree", "add", "--detach", dir, tag); err != nil {
		return problem.New(fmt.Sprintf("could not check %s out in a temporary worktree: %v", tag, err), "git worktree prune, then run it again")
	}
	defer func() {
		_, _ = g.Run("worktree", "remove", "--force", dir)
		_, _ = g.Run("worktree", "prune")
	}()
	return fn(dir)
}
