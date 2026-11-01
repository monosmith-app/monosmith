package dialog

import "slices"

// CommitLabel is the text of the box that makes a change a commit.
const CommitLabel = "Commit this change so it can be undone"

// committing lists the commands that can end in a commit: the ones that change
// the workspace and accept --commit.
var committing = []string{"adopt", "new", "add", "remove", "convert", "upgrade"}

// Committable is true for a command that accepts --commit, so its confirmation
// offers the box. A command that already carries the flag is not offered it twice.
func Committable(argv []string) bool {
	return CanCommit(argv) && !slices.Contains(argv, "--commit")
}

// CanCommit is true for a command that accepts --commit, whether or not the
// line carries it.
func CanCommit(argv []string) bool {
	return len(argv) > 0 && slices.Contains(committing, argv[0])
}

// WithCommit returns the command line with --commit added after the verb, or
// as it is when commit is false or the command has the flag already.
func WithCommit(argv []string, commit bool) []string {
	if !commit || len(argv) == 0 || slices.Contains(argv, "--commit") {
		return argv
	}
	return append(slices.Clone(argv), "--commit")
}
