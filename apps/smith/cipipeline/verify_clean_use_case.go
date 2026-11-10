package cipipeline

import (
	"errors"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
)

// uncommitted lists the tracked files changed in the working tree, as the
// release's preflight sees them; nil when git cannot say.
func uncommitted(s *session.Session) []string {
	out, err := s.Git.Run("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return nil
	}
	return release.Uncommitted(out)
}

// verifyClean fails verify when its tasks left tracked files changed: the
// release refuses a dirty tree, and the pull request is where to find out.
// A file already changed before verify fails it only in CI, where checkout
// starts clean; on a laptop it is the developer's own edit.
func verifyClean(s *session.Session, before []string, summary taskrun.Summary) error {
	after := uncommitted(s)
	if len(after) == 0 {
		return nil
	}
	ran := make([]string, 0, len(summary.Results))
	for _, r := range summary.Results {
		ran = append(ran, r.Task)
	}
	var changes []TrackedChange
	for _, c := range TrackedChanges(before, after, s.Graph.Projects, ran) {
		if !c.Before || os.Getenv("CI") != "" {
			changes = append(changes, c)
		}
	}
	if len(changes) == 0 {
		return nil
	}
	return errors.New(ExplainTrackedChanges(changes))
}
