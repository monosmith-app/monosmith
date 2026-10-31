package operation

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// exclude keeps smith's own state out of every operation commit.
const exclude = ":(exclude).smith"

// Pending is an operation that starts from a clean work tree and will be
// committed when it ends.
type Pending struct {
	root string
	git  *gitclient.Git
}

// Begin decides whether the operation will be committed. It returns nil, and
// a reason when there is one to tell the person, unless commit_operations is
// true in smith.toml, the workspace is in a git work tree with a commit, and
// the work tree has no uncommitted change. A dirty tree is never committed:
// the commit would mix the operation with other work.
//
// Parameters: root is the workspace root.
// Returns: the pending operation or nil, and a note for the person or "".
func Begin(root string) (*Pending, string) {
	cfg, err := workspace.Load(root)
	if err != nil || !cfg.Workspace.CommitOperations {
		return nil, ""
	}
	git := gitclient.New(root)
	if !git.IsRepository() {
		return nil, ""
	}
	if _, err := git.Run("rev-parse", "--verify", "--quiet", "HEAD"); err != nil {
		return nil, "not committing this operation: the repository has no commit yet"
	}
	status, err := git.Run("status", "--porcelain", "--", ".", exclude)
	if err != nil {
		return nil, "not committing this operation: " + err.Error()
	}
	if status != "" {
		return nil, "not committing this operation: the work tree has uncommitted changes"
	}
	return &Pending{root: root, git: git}, ""
}

// Finish commits what the operation changed and records it for smith undo. It
// does nothing, and returns an empty hash, when the operation changed no file.
// The commit runs the repository's hooks. If git refuses the commit, the
// changes stay in the work tree, unstaged.
//
// Parameters: command is the human text of the operation ("add package billing").
// Returns: the commit hash, or "" when nothing changed; an error when git failed.
func (p *Pending) Finish(command string) (string, error) {
	if _, err := p.git.Run("add", "-A", "--", "."); err != nil {
		return "", err
	}
	// smith's own state never belongs to a commit, ignored or not.
	_, _ = p.git.Run("reset", "--quiet", "--", ".smith")
	staged, err := p.git.Run("diff", "--cached", "--name-only", "--relative", "--", ".", exclude)
	if err != nil {
		return "", err
	}
	if staged == "" {
		return "", nil
	}
	body := fmt.Sprintf("Made by smith: %s.\n\nsmith undo reverts this commit.", command)
	if _, err := p.git.Run("commit", "-m", "chore: "+command, "-m", body); err != nil {
		_, _ = p.git.Run("reset", "--quiet")
		return "", err
	}
	hash, err := p.git.Run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	log, err := ReadLog(p.root)
	if err != nil {
		return hash, err
	}
	id := 1
	if n := len(log.Operations); n > 0 {
		id = log.Operations[n-1].ID + 1
	}
	log.Operations = append(log.Operations, Record{
		ID: id, Command: command, Commit: hash, Files: strings.Split(staged, "\n"),
		Time: time.Now().UTC().Format(time.RFC3339),
	})
	return hash, WriteLog(p.root, log)
}
