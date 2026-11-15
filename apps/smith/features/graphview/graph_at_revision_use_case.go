package graphview

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// NodesAt discovers the project graph as it was at a revision: the revision
// is checked out into a temporary worktree (the working tree is not
// touched), opened as a workspace and discovered by the same plugins.
func NodesAt(git *gitclient.Git, rev, relDir, version string, log io.Writer) ([]Node, error) {
	commit, err := git.RevParse(rev + "^{commit}")
	if err != nil {
		return nil, fmt.Errorf("%s is not a revision: %w", rev, err)
	}
	tmp, err := os.MkdirTemp("", "smith-graph-")
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(tmp, "worktree")
	if _, err := git.Run("worktree", "add", "--detach", "--quiet", dir, commit); err != nil {
		_ = os.RemoveAll(tmp)
		return nil, err
	}
	defer func() {
		_, _ = git.Run("worktree", "remove", "--force", dir)
		_ = os.RemoveAll(tmp)
	}()
	s, err := session.Open(session.Options{Dir: filepath.Join(dir, filepath.FromSlash(relDir)), Version: version, Log: log})
	if err != nil {
		return nil, fmt.Errorf("at %s: %w", rev, err)
	}
	defer s.Close()
	return Nodes(s.Graph, nil), nil
}
