package sandbox

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// Options say how to try a command.
type Options struct {
	// Root is the real workspace root; it must be a git work tree.
	Root string
	// Command is the smith command line to try, without "smith".
	Command []string
	// Checks are run in the copy after the command, in order.
	Checks []string
	// Env is added to every process: the shared cache folder, for one.
	Env []string
	// Keep leaves the copy on disk and reports its folder.
	Keep bool
	// Run starts smith.
	Run Runner
}

// Try applies Command in a temporary copy of the work tree and runs the
// checks there. The real tree is read, never written.
//
// The copy is a detached git worktree of HEAD (no node_modules or other
// ignored folder is copied) with the uncommitted changes laid over it.
//
// Returns the result; an error when the command is not one smith try accepts,
// the folder is not a git work tree, or the copy cannot be made. A failing
// command or check is in the result, not an error.
func Try(o Options) (Result, error) {
	res := Result{Command: o.Command}
	if len(o.Command) == 0 || !slices.Contains(Allowed, o.Command[0]) {
		return res, fmt.Errorf("smith try runs one of: %s", strings.Join(Allowed, ", "))
	}
	for _, c := range o.Checks {
		if !slices.Contains(Checks, c) {
			return res, fmt.Errorf("%q is not a check: use %s", c, strings.Join(Checks, ", "))
		}
	}
	git := gitclient.New(o.Root)
	if !git.IsRepository() || !git.HasCommits() {
		return res, errors.New("smith try needs a git repository with a commit: it copies the work tree with git")
	}
	status, err := git.Run("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return res, err
	}
	dir, err := os.MkdirTemp("", "smith-try-")
	if err != nil {
		return res, err
	}
	// git wants to create the folder itself.
	if err := os.Remove(dir); err != nil {
		return res, err
	}
	if _, err := git.Run("worktree", "add", "--detach", "--quiet", dir, "HEAD"); err != nil {
		return res, fmt.Errorf("cannot make the copy: %w", err)
	}
	if o.Keep {
		res.Kept = dir
	} else {
		defer func() {
			_, _ = git.Run("worktree", "remove", "--force", dir)
			_ = os.RemoveAll(dir)
			_, _ = git.Run("worktree", "prune")
		}()
	}
	if err := overlay(o.Root, dir, parseStatus(status)); err != nil {
		return res, fmt.Errorf("cannot copy your uncommitted changes: %w", err)
	}
	copyGit := gitclient.New(dir)
	before, _ := copyGit.Run("status", "--porcelain=v1", "--untracked-files=all")
	out, code, err := o.Run(dir, o.Env, o.Command...)
	if err != nil {
		return res, err
	}
	res.CommandOK, res.CommandTail = code == 0, tail(out, 12)
	after, _ := copyGit.Run("status", "--porcelain=v1", "--untracked-files=all")
	res.Changed = newLines(before, after)
	if !res.CommandOK {
		return res, nil
	}
	for _, name := range o.Checks {
		out, code, err := o.Run(dir, o.Env, name)
		if err != nil {
			return res, err
		}
		c := CheckResult{Name: name, OK: code == 0}
		if !c.OK {
			c.Tail = tail(out, 12)
		}
		res.Checks = append(res.Checks, c)
	}
	return res, nil
}

// newLines are the lines of after that before does not have.
func newLines(before, after string) []string {
	had := map[string]bool{}
	for _, l := range strings.Split(before, "\n") {
		had[l] = true
	}
	var out []string
	for _, l := range strings.Split(after, "\n") {
		if l != "" && !had[l] {
			out = append(out, l)
		}
	}
	return out
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\r\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
