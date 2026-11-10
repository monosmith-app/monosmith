package releaseview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The values rowOf holds for a line that is a button or a fold, not a plan item.
const (
	// RowRelease is the line of the "Release now" button.
	RowRelease = -1 - iota
	// RowFirst is the line of the "Create the first release commit" button.
	RowFirst
	// RowWhy is the line of the "Why?" fold.
	RowWhy
)

// emptyHeadline is the one sentence that says why nothing releases.
func (v Model) emptyHeadline() string {
	switch {
	case len(v.Plan.Held) > 0:
		return words.Say("tui.release_empty.held", "list", strings.Join(v.Plan.Held, ", "))
	case len(v.Plan.Unchanged) == 0 && v.Plan.Channel == "":
		return words.Say("tui.release_empty.no_package")
	}
	return words.Say("tui.release_empty.nothing")
}

// emptyLines adds, under the headline of an empty plan, the closed fold "Why?"
// with the conventional-commit explanation, and the action that starts a first
// release when a package was never released. It records their rows in rowOf.
func (v Model) emptyLines(th theme.Theme, lines []string, rowOf map[int]int) []string {
	mark := th.Icons.Closed
	if v.WhyOpen {
		mark = th.Icons.Open
	}
	rowOf[len(lines)] = RowWhy
	lines = append(lines, th.Accent(words.Say("tui.release_empty.why", "mark", mark, "key", commands.ShortcutOf("rel.why"))))
	if v.WhyOpen {
		for _, k := range []string{"tui.release_empty.why_commits", "tui.release_empty.why_folder", "tui.release_empty.why_chore"} {
			lines = append(lines, "  "+words.Say(k))
		}
		lines = append(lines, "  "+words.Say("tui.release_empty.why_write", "key", commands.ShortcutOf("rel.why")))
	}
	if v.FirstRelease() && v.FirstButton != "" {
		lines = append(lines, "")
		rowOf[len(lines)] = RowFirst
		lines = append(lines, v.FirstButton, "  "+th.Bold(words.Say("tui.release_empty.first_command", "command", release.FirstReleaseCommand)), "  "+th.Muted(words.Say("tui.release_empty.first_hint")))
	}
	return lines
}
