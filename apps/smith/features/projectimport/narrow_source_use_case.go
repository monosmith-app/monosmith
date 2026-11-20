package projectimport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// pathBranch is the branch of the temporary clone that holds the history of
// the imported sub-folder only, with its files at the top.
const pathBranch = "smith-import-path"

// Narrow makes the source the sub-folder sub (--path), so the import takes
// that project only. A plain folder is read from the sub-folder. A git source
// gets the history of the sub-folder only: in a temporary clone (the source
// itself is never written), git subtree split writes each commit that
// touches sub again with sub's files at the top, and the import takes that
// branch, so --history keep and squash bring the project's own commits.
//
// Parameters: src is the fetched source; sub is the sub-folder, relative to
// the source, with forward slashes; out gets the progress.
// Returns: the narrowed source (Release removes its clone), or a problem for
// a sub-folder that is not in the source and for a split that fails.
func Narrow(src Source, sub string, out io.Writer) (Source, error) {
	clean, err := adoptmove.CleanFolder(sub)
	if err != nil {
		return src, problem.New("--path "+sub+" is not a folder inside the source", "give the project's folder inside the source, such as --path src/MyLib")
	}
	if info, err := os.Stat(filepath.Join(src.Dir, filepath.FromSlash(clean))); err != nil || !info.IsDir() {
		return src, problem.New(src.Arg+" has no folder "+clean, "give the project's folder inside the source, such as --path src/MyLib")
	}
	src.Path = clean
	if !src.Git {
		src.Dir = filepath.Join(src.Dir, filepath.FromSlash(clean))
		return src, nil
	}
	commit, err := gitclient.New(src.Dir).Run("rev-parse", src.Ref+"^{commit}")
	if err != nil {
		return src, fmt.Errorf("the commit of %s could not be read: %s", src.Ref, firstLine(err.Error()))
	}
	if src.Temp == "" {
		temp, err := os.MkdirTemp("", "smith-import-")
		if err != nil {
			return src, err
		}
		if _, err := gitclient.New(temp).Run("clone", "--quiet", filepath.ToSlash(src.Dir), "source"); err != nil {
			_ = os.RemoveAll(temp)
			return src, fmt.Errorf("the source could not be cloned for the split: %s", firstLine(err.Error()))
		}
		src.Temp, src.Dir = temp, filepath.Join(temp, "source")
	}
	git := gitclient.New(src.Dir)
	fmt.Fprintf(out, "… git subtree split --prefix=%s (the commits of %s only)\n", clean, clean)
	head, err := git.Run("subtree", "split", "--quiet", "--prefix="+clean, commit)
	if err != nil {
		return src, problem.New("git subtree split could not take the history of "+clean+": "+firstLine(err.Error()),
			"use --history copy to copy the files of "+clean+" without the history")
	}
	if _, err := git.Run("checkout", "--quiet", "-B", pathBranch, strings.TrimSpace(head)); err != nil {
		return src, fmt.Errorf("the history of %s could not be checked out: %s", clean, firstLine(err.Error()))
	}
	src.PathRef, src.Ref = src.Ref, pathBranch
	src.Commit, _ = git.Run("rev-parse", "--short", pathBranch)
	if n, err := git.Run("rev-list", "--count", pathBranch); err == nil {
		src.Commits, _ = strconv.Atoi(n)
	}
	src.Dirty = false
	return src, nil
}

// FolderName is the name the source's folder gives a project that has no
// name in a manifest: the last part of --path, else of the git URL (without
// .git), else of the folder.
//
// Parameters: src is the source.
// Returns: the name, as typed (adoptmove.DeriveName makes it a project name).
func FolderName(src Source) string {
	switch {
	case src.Path != "":
		return filepath.Base(filepath.FromSlash(src.Path))
	case src.URL != "":
		url := strings.TrimSuffix(strings.TrimRight(src.URL, "/"), ".git")
		return url[strings.LastIndexAny(url, "/:")+1:]
	}
	return filepath.Base(src.Dir)
}
