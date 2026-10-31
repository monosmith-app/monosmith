package operation

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// Candidate is the operation smith undo would revert, once it passed every check.
type Candidate struct {
	Record  Record
	Changes []operationplan.Line
}

// Plan is what the revert will do: the inverse of each change the operation made.
func (c Candidate) Plan() operationplan.Plan {
	p := operationplan.Plan{Title: "undo " + c.Record.Command + " (commit " + short(c.Record.Commit) + ")"}
	for _, l := range c.Changes {
		p.Add("Reverts the operation", l)
	}
	return p
}

func short(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}

// Prepare finds the last operation smith made and checks that it can be
// reverted safely. It refuses, with the reason and the next step, when no
// operation is recorded, the commit is gone from the history (it was rewritten),
// a later commit changed one of its files, or a file of it has an uncommitted
// change. It changes nothing.
//
// Parameters: root is the workspace root.
// Returns: the candidate, or a *problem.Error for a refusal.
func Prepare(root string) (Candidate, error) {
	log, err := ReadLog(root)
	if err != nil {
		return Candidate{}, err
	}
	last := log.Last()
	if last == nil {
		return Candidate{}, problem.New("smith has no operation to undo here", "set commit_operations = true under [workspace] in smith.toml, then run a command such as smith add")
	}
	git := gitclient.New(root)
	if !git.IsRepository() {
		return Candidate{}, problem.New("this workspace is not in a git work tree", "smith undo reverts a commit; use git to undo the change")
	}
	if _, err := git.Run("cat-file", "-e", last.Commit+"^{commit}"); err != nil {
		return Candidate{}, problem.New("commit "+short(last.Commit)+" (smith "+last.Command+") is not in the repository", "the history was rewritten; undo the change with git")
	}
	if _, err := git.Run("merge-base", "--is-ancestor", last.Commit, "HEAD"); err != nil {
		return Candidate{}, problem.New("commit "+short(last.Commit)+" (smith "+last.Command+") is not in the history of the current branch", "check out the branch that holds it, or undo the change with git")
	}
	since, err := git.Run("diff", "--name-only", "--relative", last.Commit, "HEAD", "--")
	if err != nil {
		return Candidate{}, err
	}
	if touched := intersect(last.Files, since); len(touched) > 0 {
		return Candidate{}, problem.New("files changed since smith "+last.Command+": "+strings.Join(touched, ", "), "undo the later changes first, or revert the commit with git and resolve the conflicts yourself")
	}
	args := append([]string{"status", "--porcelain", "--"}, last.Files...)
	if dirty, err := git.Run(args...); err != nil {
		return Candidate{}, err
	} else if dirty != "" {
		return Candidate{}, problem.New("files of smith "+last.Command+" have uncommitted changes", "commit or discard them, then run smith undo again")
	}
	status, err := git.Run("diff-tree", "--no-commit-id", "--name-status", "-r", "--no-renames", "--relative", last.Commit)
	if err != nil {
		return Candidate{}, err
	}
	c := Candidate{Record: *last}
	for _, row := range strings.Split(status, "\n") {
		code, path, ok := strings.Cut(row, "\t")
		if !ok {
			continue
		}
		switch code {
		case "A":
			c.Changes = append(c.Changes, operationplan.Line{Action: operationplan.Remove, Path: path})
		case "D":
			c.Changes = append(c.Changes, operationplan.Line{Action: operationplan.Create, Path: path})
		default:
			c.Changes = append(c.Changes, operationplan.Line{Action: operationplan.Change, Path: path})
		}
	}
	return c, nil
}

func intersect(files []string, changed string) []string {
	set := map[string]bool{}
	for _, f := range strings.Split(changed, "\n") {
		set[f] = true
	}
	var out []string
	for _, f := range files {
		if set[f] {
			out = append(out, f)
		}
	}
	return out
}

// Undo reverts the candidate with git revert: a new commit, with history left
// as it is. If the revert does not apply cleanly, it is aborted and the work
// tree is as it was.
//
// Parameters: root is the workspace root; c comes from Prepare.
// Returns: the hash of the revert commit.
func Undo(root string, c Candidate) (string, error) {
	git := gitclient.New(root)
	if _, err := git.Run("revert", "--no-commit", c.Record.Commit); err != nil {
		_, _ = git.Run("revert", "--abort")
		return "", problem.New("git could not revert commit "+short(c.Record.Commit)+": "+err.Error(), "resolve it with git revert "+short(c.Record.Commit))
	}
	body := fmt.Sprintf("This reverts commit %s, made by smith: %s.", c.Record.Commit, c.Record.Command)
	if _, err := git.Run("commit", "-m", "chore: undo "+c.Record.Command, "-m", body); err != nil {
		_, _ = git.Run("revert", "--abort")
		return "", err
	}
	hash, err := git.Run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	log, err := ReadLog(root)
	if err != nil {
		return hash, err
	}
	for i := range log.Operations {
		if log.Operations[i].ID == c.Record.ID {
			log.Operations[i].Undone = true
			log.Operations[i].UndoneBy = hash
		}
	}
	return hash, WriteLog(root, log)
}
