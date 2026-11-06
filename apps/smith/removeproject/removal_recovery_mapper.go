package removeproject

import "fmt"

// RecoveryLines say how to get a removed project back. smith removes the
// folder and edits the workspace files, so the way back depends on what was
// kept: a copy under .smith/removed, or git's last commit. Nothing here runs.
//
// Parameters: r is the removal that happened (not a dry run).
// Returns: the lines to print, each starting with two spaces.
func RecoveryLines(r Removal) []string {
	lines := []string{"  To get it back:"}
	if r.KeptAt != "" {
		lines = append(lines, fmt.Sprintf("    move %s back to %s (a copy of the whole folder)", r.KeptAt, r.Root))
	}
	lines = append(lines,
		fmt.Sprintf("    git restore --source=HEAD --staged --worktree -- %s   (the files as of the last commit; untracked files are not in git)", r.Root),
		"    git diff shows the other edits of this removal (smith.toml, the pipelines); git restore <file> undoes one",
		"    smith undo reverts it when it was committed (smith remove --commit, or smith commit)",
	)
	return lines
}
