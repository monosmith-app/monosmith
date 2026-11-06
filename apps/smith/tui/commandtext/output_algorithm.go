package commandtext

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// DryRunLine ends the plan a command prints with --dry-run.
func DryRunLine() string { return words.Say("tui.capture.dry_run") }

// WithoutDryRunLine drops the closing "Dry run" line of a plan: the dialog
// above it already says that nothing is written yet, and a first-time user
// does not know the words "dry run".
//
// Parameters: lines are the lines of the plan.
// Returns: the lines without the closing "Dry run" line.
func WithoutDryRunLine(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == DryRunLine() {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// Lines splits the output of a command into lines and drops the blank ones at
// the ends.
func Lines(text string) []string {
	return strings.Split(strings.Trim(text, "\n"), "\n")
}

// Failure is what to show for a command that failed: its output, or why it
// never ran (err), or a plain sentence when it gave neither.
func Failure(text string, err error) string {
	switch {
	case strings.TrimSpace(text) != "":
		return text
	case err != nil:
		return err.Error()
	}
	return words.Say("tui.capture.the_command_ended_with_an_error")
}

// FirstLine is the first non-blank line of a text, trimmed.
func FirstLine(text string) string {
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}
