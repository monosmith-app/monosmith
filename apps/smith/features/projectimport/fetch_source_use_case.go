package projectimport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// Fetch makes the source readable: a local folder as it is, a git URL as a
// clone in a temporary folder. For a git repository it settles the ref (the
// one asked for, else the branch that is checked out) and counts its commits.
// The workspace is not changed. Release removes the clone.
//
// Parameters: arg is the source as typed; ref is --ref or ""; out gets the
// progress line of a clone.
// Returns: the source, or a problem for a source that is neither a folder nor
// a git URL, a clone that fails, or a ref the repository does not have.
func Fetch(arg, ref string, out io.Writer) (Source, error) {
	src := Source{Arg: arg}
	if info, err := os.Stat(arg); err == nil && info.IsDir() {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return src, err
		}
		src.Dir = abs
		git := gitclient.New(abs)
		top, err := git.TopLevel()
		if err != nil || !sameFolder(top, abs) || !git.HasCommits() {
			if ref != "" {
				return src, problem.New("--ref "+ref+" needs a git repository, and "+arg+" is a plain folder", "drop --ref: the folder is copied as it is")
			}
			return src, nil
		}
		status, _ := git.Run("status", "--porcelain")
		src.Dirty = strings.TrimSpace(status) != ""
		return settleRef(src, ref)
	}
	if !IsGitURL(arg) {
		return src, problem.New(arg+" is neither a folder nor a git URL", "give a folder (smith import ../my-app) or a git URL (smith import https://github.com/owner/repo)")
	}
	temp, err := os.MkdirTemp("", "smith-import-")
	if err != nil {
		return src, err
	}
	src.URL, src.Temp, src.Dir = arg, temp, filepath.Join(temp, "source")
	fmt.Fprintf(out, "… cloning %s\n", arg)
	if _, err := gitclient.New(temp).Run("clone", "--quiet", arg, "source"); err != nil {
		Release(src)
		return Source{}, problem.New("git could not clone "+arg+": "+firstLine(err.Error()), "check the URL and your access to it (git clone "+arg+" must work), then run smith import again")
	}
	if ref != "" {
		if _, err := gitclient.New(src.Dir).Run("checkout", "--quiet", ref); err != nil {
			Release(src)
			return Source{}, problem.New(arg+" has no branch, tag or commit "+ref, "drop --ref to import the default branch, or name one that git ls-remote "+arg+" lists")
		}
	}
	return settleRef(src, ref)
}

// settleRef fills the ref, its commit and the number of commits of a git source.
func settleRef(src Source, ref string) (Source, error) {
	git := gitclient.New(src.Dir)
	src.Git = true
	if ref == "" {
		ref = git.CurrentBranch()
		if ref == "" || ref == "HEAD" {
			ref, _ = git.Run("rev-parse", "HEAD")
		}
	}
	commit, err := git.Run("rev-parse", "--short", ref+"^{commit}")
	if err != nil {
		return src, problem.New(src.Arg+" has no branch, tag or commit "+ref, "name one that git branch -a or git tag lists in the source")
	}
	src.Ref, src.Commit = ref, commit
	if n, err := git.Run("rev-list", "--count", ref); err == nil {
		src.Commits, _ = strconv.Atoi(n)
	}
	return src, nil
}

// Release removes the temporary clone of a source, if there is one.
//
// Parameters: src is the source Fetch returned.
func Release(src Source) {
	if src.Temp != "" {
		_ = os.RemoveAll(src.Temp)
	}
}

func sameFolder(a, b string) bool {
	return strings.EqualFold(filepath.Clean(filepath.FromSlash(a)), filepath.Clean(b))
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}
