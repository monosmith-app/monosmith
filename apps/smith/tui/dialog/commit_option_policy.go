package dialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"slices"
)

// CommitLabel is the text of the box that makes a change a commit.
var CommitLabel = words.Say("dialog.commit.label")

// committing lists the commands that can end in a commit: the ones that change
// the workspace and accept --commit. A command of two words is written with
// both ("config set").
var committing = []string{
	"adopt", "new", "add", "remove", "convert", "upgrade", "swap", "derive",
	"install", "uninstall", "sync", "up", "kind", "bootstrap", "architecture",
	"environments add", "environments remove", "tauri add", // words:literal
	"config set", "config unset", "plugin enable", "plugin disable", // words:literal
}

// Committable is true for a command that accepts --commit, so its confirmation
// offers the box. A command that already carries the flag is not offered it twice.
func Committable(argv []string) bool {
	return CanCommit(argv) && !slices.Contains(argv, "--commit")
}

// CanCommit is true for a command that accepts --commit, whether or not the
// line carries it.
func CanCommit(argv []string) bool {
	if len(argv) == 0 {
		return false
	}

	return slices.Contains(committing, argv[0]) || (len(argv) > 1 && slices.Contains(committing, argv[0]+" "+argv[1]))
}

// WithCommit returns the command line with --commit added after the verb, or
// as it is when commit is false or the command has the flag already.
func WithCommit(argv []string, commit bool) []string {
	if !commit || len(argv) == 0 || slices.Contains(argv, "--commit") {
		return argv
	}
	return append(slices.Clone(argv), "--commit")
}
