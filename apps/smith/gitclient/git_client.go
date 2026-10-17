package gitclient

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Git runs git in a repository.
type Git struct {
	Dir string
}

// New returns a client for the repository containing dir.
func New(dir string) *Git { return &Git{Dir: dir} }

// Run executes git with args and returns trimmed stdout.
func (g *Git) Run(args ...string) (string, error) {
	out, err := g.raw(args...)
	return strings.TrimRight(out, "\r\n"), err
}

// RunWithEnv is Run with extra environment variables. A secret goes here,
// never in args: an error message repeats the arguments.
func (g *Git) RunWithEnv(env []string, args ...string) (string, error) {
	out, err := g.rawEnv(env, args...)
	return strings.TrimRight(out, "\r\n"), err
}

func (g *Git) raw(args ...string) (string, error) { return g.rawEnv(nil, args...) }

func (g *Git) rawEnv(env []string, args ...string) (string, error) {
	full := append([]string{"-c", "core.quotepath=off", "--no-optional-locks"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = g.Dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.String(), fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// IsRepository reports whether Dir is inside a git work tree.
func (g *Git) IsRepository() bool {
	out, err := g.Run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// Files lists tracked files and untracked files that are not ignored,
// workspace-relative with forward slashes, sorted.
func (g *Git) Files() ([]string, error) {
	out, err := g.raw("ls-files", "-z", "--cached", "--others", "--exclude-standard", "--deduplicate")
	if err != nil {
		return nil, err
	}
	files := splitNUL(out)
	// A tracked file deleted from the working tree is still in the index;
	// it is not part of the workspace any more.
	if deleted, err := g.raw("ls-files", "-z", "--deleted"); err == nil && deleted != "" {
		gone := map[string]bool{}
		for _, f := range splitNUL(deleted) {
			gone[f] = true
		}
		kept := files[:0]
		for _, f := range files {
			if !gone[f] {
				kept = append(kept, f)
			}
		}
		files = kept
	}
	sort.Strings(files)
	return files, nil
}

// ChangedSince lists files changed between base and HEAD plus uncommitted
// and untracked changes.
func (g *Git) ChangedSince(base string) ([]string, error) {
	seen := map[string]bool{}
	for _, args := range [][]string{
		{"diff", "--name-only", "--no-renames", "--relative", "-z", base, "HEAD"},
		{"diff", "--name-only", "--no-renames", "--relative", "-z", "HEAD"},
		{"ls-files", "-z", "--others", "--exclude-standard"},
	} {
		out, err := g.raw(args...)
		if err != nil {
			return nil, err
		}
		for _, f := range splitNUL(out) {
			seen[f] = true
		}
	}
	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	sort.Strings(files)
	return files, nil
}

// MergeBase returns the best common ancestor of two revisions.
func (g *Git) MergeBase(a, b string) (string, error) { return g.Run("merge-base", a, b) }

// RevParse resolves a revision to a commit hash.
func (g *Git) RevParse(rev string) (string, error) {
	return g.Run("rev-parse", "--verify", "--quiet", rev+"^{commit}")
}

// IsShallow reports whether the clone has truncated history.
func (g *Git) IsShallow() bool {
	out, err := g.Run("rev-parse", "--is-shallow-repository")
	return err == nil && out == "true"
}

// CurrentBranch returns the checked-out branch, or "" when detached.
func (g *Git) CurrentBranch() string {
	out, err := g.Run("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return out
}

// HasCommits reports whether HEAD exists.
func (g *Git) HasCommits() bool {
	_, err := g.RevParse("HEAD")
	return err == nil
}

// Tags lists tags starting with prefix.
func (g *Git) Tags(prefix string) ([]string, error) {
	out, err := g.Run("tag", "--list", prefix+"*")
	if err != nil || out == "" {
		return nil, err
	}
	return strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n"), nil
}

// TagCommit returns the commit a tag points to, or "" if the tag is absent.
func (g *Git) TagCommit(tag string) string {
	out, err := g.RevParse("refs/tags/" + tag)
	if err != nil {
		return ""
	}
	return out
}

// Commit is one commit's message parts.
type Commit struct {
	Hash, Author, Subject, Body string
}

// Log returns commits reachable from HEAD and not from since (all history when
// since is empty) that touch path ("" means the whole repository), newest first.
func (g *Git) Log(since string, paths ...string) ([]Commit, error) {
	args := []string{"log", "-z", "--format=%H%x1f%an%x1f%s%x1f%b"}
	if since != "" {
		args = append(args, since+"..HEAD")
	} else {
		args = append(args, "HEAD")
	}
	var only []string
	for _, p := range paths {
		if p != "" {
			only = append(only, p)
		}
	}
	if len(only) > 0 {
		args = append(append(args, "--"), only...)
	}
	out, err := g.raw(args...)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, record := range splitNUL(out) {
		parts := strings.SplitN(record, "\x1f", 4)
		if len(parts) < 3 {
			continue
		}
		c := Commit{Hash: strings.TrimSpace(parts[0]), Author: parts[1], Subject: parts[2]}
		if len(parts) == 4 {
			c.Body = strings.TrimSpace(parts[3])
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// CreateTag creates an annotated tag at HEAD.
func (g *Git) CreateTag(name, message string) error {
	_, err := g.Run("tag", "--annotate", name, "--message", message)
	return err
}

// CreateTagAt creates an annotated tag on a given commit.
func (g *Git) CreateTagAt(name, message, commit string) error {
	_, err := g.Run("tag", "--annotate", name, "--message", message, commit)
	return err
}

// HasRemote reports whether a remote is configured.
func (g *Git) HasRemote(name string) bool {
	_, err := g.Run("remote", "get-url", name)
	return err == nil
}

// RemoteURL returns a remote's URL.
func (g *Git) RemoteURL(name string) (string, error) { return g.Run("remote", "get-url", name) }

// PushTag pushes one tag to a remote.
func (g *Git) PushTag(remote, tag string) error {
	_, err := g.Run("push", "--no-verify", remote, "refs/tags/"+tag)
	return err
}

// RestoreFiles discards working-tree changes to paths.
func (g *Git) RestoreFiles(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := g.Run(append([]string{"checkout", "--"}, paths...)...)
	return err
}

// Init creates a repository with main as its first branch.
func (g *Git) Init() error {
	_, err := g.Run("init", "--initial-branch=main")
	return err
}

// Prefix is the directory's path inside its repository: "" at the top level,
// "sub/dir/" below it.
func (g *Git) Prefix() (string, error) { return g.Run("rev-parse", "--show-prefix") }

// TopLevel is the root of the repository the directory is in.
func (g *Git) TopLevel() (string, error) { return g.Run("rev-parse", "--show-toplevel") }

// CommitPaths stages and commits only paths (relative to the directory),
// leaving every other change, staged or not, as it was.
func (g *Git) CommitPaths(message string, paths []string) error {
	if len(paths) == 0 {
		return errors.New("nothing to commit")
	}
	if _, err := g.Run(append([]string{"add", "--"}, paths...)...); err != nil {
		return err
	}
	_, err := g.Run(append([]string{"commit", "--quiet", "--message", message, "--"}, paths...)...)
	return err
}

// CommitAll stages everything and commits it.
func (g *Git) CommitAll(message string) error {
	if _, err := g.Run("add", "--all"); err != nil {
		return err
	}
	_, err := g.Run("commit", "--quiet", "--message", message)
	return err
}

func splitNUL(s string) []string {
	var out []string
	for _, part := range strings.Split(s, "\x00") {
		part = strings.TrimLeft(part, "\n")
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
