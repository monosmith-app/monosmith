package operation

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/spf13/cobra"
)

// Leftover is a folder an undone operation created that still holds generated
// files git ignores (node_modules, dist, a venv). git revert cannot remove
// them, so the folder looks half removed.
type Leftover struct {
	// Dir is the workspace-relative folder, with forward slashes.
	Dir string
	// Entries are the ignored entries inside it, such as node_modules/ and dist/.
	Entries []string
}

// FindLeftovers lists the folders the undone operation created that now hold
// nothing but ignored files. A folder that holds a tracked file, or an
// untracked file git does not ignore, is the person's and is never listed.
//
// Parameters: root is the workspace root; removed are the paths the revert
// deleted (the files the operation had added).
// Returns: the leftovers, sorted by folder; nothing is changed.
func FindLeftovers(root string, removed []string) []Leftover {
	git := gitclient.New(root)
	candidates := map[string]bool{}
	// ours says a folder holds nothing of git's and nothing of the person's.
	ours := func(dir string) bool {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil || !info.IsDir() {
			return false
		}
		if tracked, _ := git.Run("ls-files", "--", dir); tracked != "" {
			return false
		}
		own, _ := git.Run("ls-files", "--others", "--exclude-standard", "--", dir)
		return own == ""
	}
	for _, file := range removed {
		// The highest folder of the path that holds only ignored files, but never a
		// folder above the project level (apps/ or libs/ stay).
		parts := strings.Split(file, "/")
		floor := min(2, len(parts)-1)
		best := ""
		for i := len(parts) - 1; i >= floor; i-- {
			dir := path.Join(parts[:i]...)
			if !ours(dir) {
				break
			}
			best = dir
		}
		if best != "" {
			candidates[best] = true
		}
	}
	var out []Leftover
	for dir := range candidates {
		ignored, err := git.Run("ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "--", dir)
		if err != nil || ignored == "" {
			continue
		}
		var entries []string
		for _, e := range strings.Split(ignored, "\n") {
			if entry := strings.TrimPrefix(e, dir+"/"); entry != "" {
				entries = append(entries, entry)
			}
		}
		out = append(out, Leftover{Dir: dir, Entries: entries})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out
}

// settleLeftovers says what an undo left behind and, as the person chooses,
// deletes it: --clean deletes, a terminal asks, anything else only reports and
// names --clean. Nothing is deleted unasked.
//
// Parameters: cmd is the undo command; root the workspace root; left what
// FindLeftovers found; clean is --clean; yes is --yes.
// Returns: the error from deleting, if any.
func settleLeftovers(cmd *cobra.Command, root string, left []Leftover, clean, yes bool) error {
	if len(left) == 0 {
		return nil
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Left behind: git ignores these generated files, so the revert could not remove them.")
	for _, l := range left {
		fmt.Fprintln(out, leftoverLine(l))
	}
	remove := clean
	if !clean && operationplan.ShouldAsk(operationplan.DetectEnvironment(), yes) {
		fmt.Fprint(out, "Delete them? They are rebuilt by smith install and smith build. [y/N] ")
		answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		remove = strings.EqualFold(strings.TrimSpace(answer), "y")
	}
	if !remove {
		fmt.Fprintln(out, "Kept them. smith undo --clean deletes them next time, or delete the folders yourself.")
		return nil
	}
	if err := RemoveLeftovers(root, left); err != nil {
		return err
	}
	fmt.Fprintln(out, "Deleted them.")
	return nil
}

// leftoverLine is one line of the report: the folder and, when known, the
// ignored entries in it, without an empty entry or a stray comma.
func leftoverLine(l Leftover) string {
	var kept []string
	for _, e := range l.Entries {
		if e != "" {
			kept = append(kept, e)
		}
	}
	if len(kept) == 0 {
		return fmt.Sprintf("  %s/", l.Dir)
	}
	return fmt.Sprintf("  %s/  (%s)", l.Dir, strings.Join(kept, ", "))
}

// RemoveLeftovers deletes the folders FindLeftovers listed. They hold only
// generated files that git ignores, so the next install or build makes them again.
//
// Parameters: root is the workspace root; left are the folders to delete.
// Returns: the first error from the file system.
func RemoveLeftovers(root string, left []Leftover) error {
	for _, l := range left {
		if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(l.Dir))); err != nil {
			return err
		}
	}
	return nil
}
