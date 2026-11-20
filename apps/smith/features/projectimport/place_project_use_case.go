package projectimport

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptsplit"
	"github.com/monosmith-app/monosmith/apps/smith/features/lockcarry"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// skippedFolders are never copied from a plain folder: installed
// dependencies, virtual environments, caches and git's own folder.
var skippedFolders = map[string]bool{
	".git": true, "node_modules": true, ".venv": true, "venv": true, "__pycache__": true,
	".pytest_cache": true, ".mypy_cache": true, ".ruff_cache": true, ".tox": true,
}

// Place puts the project into its folder of the workspace at root. keep
// merges the source's commits, written again under the folder (one merge
// commit); squash runs git subtree add --squash (two commits); copy copies
// the files and commits nothing. keep and squash need a clean work tree:
// the caller checks.
//
// Parameters: root is the workspace root; p is the settled plan; out gets the
// progress.
// Returns: an error from git or from the copy; a failed copy leaves the
// folder for the person to remove (the message says so).
func Place(root string, p Plan, out io.Writer) error {
	if p.History == HistoryCopy {
		n, err := copyProject(p.Source, filepath.Join(root, filepath.FromSlash(p.Folder)))
		if err != nil {
			return fmt.Errorf("the copy into %s stopped (%w); remove %s, then run smith import again", p.Folder, err, p.Folder)
		}
		fmt.Fprintf(out, "  copied %d files into %s\n", n, p.Folder)
		return nil
	}
	if p.History == HistoryKeep {
		n, err := keepHistory(root, p, out)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "  placed %s with its %d commit(s): git log -- %s shows them\n", p.Folder, n, p.Folder)
		return nil
	}
	fmt.Fprintf(out, "… git subtree add --prefix=%s --squash\n", p.Folder)
	if _, err := gitclient.New(root).Run("subtree", "add", "--prefix="+p.Folder, "--squash", "-m", CommitMessage(p), filepath.ToSlash(p.Source.Dir), p.Source.Ref); err != nil {
		return fmt.Errorf("git subtree add failed: %s", firstLine(err.Error()))
	}
	fmt.Fprintf(out, "  placed %s as one squashed commit of %s (%s)\n", p.Folder, p.Source.Ref, p.Source.Commit)
	return nil
}

// historyBranch is the branch of the temporary clone that holds the history
// rewritten into the project's folder.
const historyBranch = "refs/heads/smith-import"

// keepHistory brings every commit of the source into the workspace under the
// project's folder. In a bare clone of the source (the source itself is
// never written), git fast-import writes each commit again with its files
// under the folder; the workspace then fetches that branch and merges it,
// as one merge commit, with --allow-unrelated-histories. Unlike git subtree
// add, the commits hold the folder's paths, so git log -- <folder> and git
// blame follow them. It returns the number of commits.
func keepHistory(root string, p Plan, out io.Writer) (int, error) {
	temp, err := os.MkdirTemp("", "smith-history-")
	if err != nil {
		return 0, err
	}
	defer func() { _ = os.RemoveAll(temp) }()
	fmt.Fprintf(out, "… writing the %d commit(s) of %s into %s\n", p.Source.Commits, p.Source.Ref, p.Folder)
	if _, err := gitclient.New(temp).Run("clone", "--quiet", "--bare", filepath.ToSlash(p.Source.Dir), "history"); err != nil {
		return 0, fmt.Errorf("the history of the source could not be read: %s", firstLine(err.Error()))
	}
	clone := gitclient.New(filepath.Join(temp, "history"))
	log, err := clone.Run("log", "--reverse", "--topo-order", "--date=raw", "--format="+LogFormat, p.Source.Ref)
	if err != nil {
		return 0, fmt.Errorf("the history of %s could not be read: %s", p.Source.Ref, firstLine(err.Error()))
	}
	stream, n := HistoryStream(log, historyBranch, p.Folder)
	if _, err := clone.RunInput(stream, "fast-import", "--quiet"); err != nil {
		return 0, fmt.Errorf("the history could not be written into %s: %s", p.Folder, firstLine(err.Error()))
	}
	ws := gitclient.New(root)
	if _, err := ws.Run("fetch", "--quiet", "--no-tags", filepath.ToSlash(filepath.Join(temp, "history")), historyBranch); err != nil {
		return 0, fmt.Errorf("the workspace could not fetch the history: %s", firstLine(err.Error()))
	}
	if _, err := ws.Run("merge", "--quiet", "--allow-unrelated-histories", "--no-ff", "-m", CommitMessage(p), "FETCH_HEAD"); err != nil {
		return 0, fmt.Errorf("git merge of the history failed (%s); git merge --abort takes it back", firstLine(err.Error()))
	}
	return n, nil
}

// CommitMessage is the message of the merge commit the import makes: the
// project, the source and the commit it was taken from.
//
// Parameters: p is the plan.
// Returns: the message.
func CommitMessage(p Plan) string {
	from := p.Source.URL
	if from == "" {
		from = p.Source.Arg
	}
	text := fmt.Sprintf("chore: import %s into %s\n\nImported by smith import from %s at %s (%s)", p.Name, p.Folder, from, p.Source.Ref, p.Source.Commit)
	if p.History == HistorySquash {
		text += ", squashed"
	}
	return text + "."
}

// copyProject copies the source into dst and returns the number of files. A
// git source copies the files git tracks or does not ignore; a plain folder
// copies everything but skippedFolders.
func copyProject(src Source, dst string) (int, error) {
	var files []string
	if src.Git {
		list, err := gitclient.New(src.Dir).Files()
		if err != nil {
			return 0, err
		}
		files = list
	} else {
		err := filepath.WalkDir(src.Dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != src.Dir && skippedFolders[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(src.Dir, p)
			if err == nil && d.Type().IsRegular() {
				files = append(files, filepath.ToSlash(rel))
			}
			return err
		})
		if err != nil {
			return 0, err
		}
	}
	for _, rel := range files {
		raw, err := os.ReadFile(filepath.Join(src.Dir, filepath.FromSlash(rel)))
		if err != nil {
			return 0, err
		}
		to := filepath.Join(dst, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return 0, err
		}
		if err := os.WriteFile(to, raw, 0o644); err != nil {
			return 0, err
		}
	}
	return len(files), nil
}

// DropLockfiles removes the project's own lockfiles from the top of its
// folder and returns their names.
//
// Parameters: root is the workspace root; folder is the project's; patterns
// are the lockfile names (Lockfiles).
// Returns: the names removed, or an error from the file system.
func DropLockfiles(root, folder string, patterns []string) ([]string, error) {
	dir := filepath.Join(root, filepath.FromSlash(folder))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dropped []string
	for _, e := range entries {
		if e.IsDir() || !matches(patterns, strings.ToLower(e.Name())) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			return dropped, err
		}
		dropped = append(dropped, e.Name())
	}
	return dropped, nil
}

// LockfilesIn names the lockfiles at the top of a source folder, for the plan.
//
// Parameters: dir is the source folder; patterns are the lockfile names.
// Returns: the names found.
func LockfilesIn(dir string, patterns []string) []string {
	entries, _ := os.ReadDir(dir)
	var out []string
	for _, e := range entries {
		if !e.IsDir() && matches(patterns, strings.ToLower(e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out
}

// shownVersions is how many kept versions KeepVersions names.
const shownVersions = 8

// KeepVersions keeps the versions the project's own lockfile locked, before
// the first install of the workspace resolves its ranges again. For a Node
// project the npm lockfile of the source goes into the workspace's
// package-lock.json under the project's folder (lockcarry). A lockfile smith
// cannot carry is an item left, so the person knows the versions can move.
//
// Parameters: root is the workspace root; p is the plan, whose source folder
// still holds the lockfile; out gets the versions kept.
// Returns: the items left, or nil.
func KeepVersions(root string, p Plan, out io.Writer) []Left {
	if p.Ecosystem == adoptsplit.Python {
		for _, l := range p.Lockfiles {
			if strings.EqualFold(l, "uv.lock") {
				return []Left{{
					Problem: "the versions " + l + " locked are not kept: the workspace installs " + p.Name + " with pip, which reads the ranges of pyproject.toml again",
					Remedy:  "if a newer version breaks " + p.Name + ", pin it in " + p.Folder + "/pyproject.toml, then run smith install -p " + p.Name,
				}}
			}
		}
		return nil
	}
	if p.Ecosystem != "node" {
		return nil
	}
	r, err := lockcarry.CarryNpm(root, p.Source.Dir, p.Folder)
	switch {
	case err != nil:
		return []Left{{
			Problem: "the versions the project's lockfile locked are not kept: " + firstLine(err.Error()),
			Remedy:  "if a newer version breaks " + p.Name + ", pin it in " + p.Folder + "/package.json, then run smith install -p " + p.Name,
		}}
	case r.Unsupported != "":
		return []Left{{
			Problem: "the versions " + r.Unsupported + " locked are not kept: smith reads npm lockfiles only, so npm resolves the ranges of package.json again",
			Remedy:  "if a newer version breaks " + p.Name + ", pin the version " + r.Unsupported + " names in " + p.Folder + "/package.json, then run smith install -p " + p.Name,
		}}
	case r.NoWorkspaceLock:
		fmt.Fprintf(out, "  split: the workspace has no package-lock.json yet, so the install resolves the ranges of %s again\n", p.Name)
	case r.Lockfile != "":
		fmt.Fprintf(out, "  split: kept the versions %s locked (%d packages)%s\n", r.Lockfile, r.Added, KeptList(r.Kept))
	}
	return nil
}

// KeptList names the direct dependencies and their kept versions, at most
// shownVersions of them.
//
// Parameters: kept are the direct dependencies and their versions.
// Returns: ": name version, ..." or "" when kept is empty.
func KeptList(kept []lockcarry.Locked) string {
	if len(kept) == 0 {
		return ""
	}
	var names []string
	for i, k := range kept {
		if i == shownVersions {
			names = append(names, fmt.Sprintf("and %d more", len(kept)-shownVersions))
			break
		}
		names = append(names, k.Name+" "+k.Version)
	}
	return ": " + strings.Join(names, ", ")
}

func matches(patterns []string, name string) bool {
	for _, p := range patterns {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}
