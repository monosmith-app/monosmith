package releaseconfirm

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Result is the dialog that ends a release: what happened, the new tags and a
// line that says what to do next.
//
// Parameters: ok is true when the release ended with exit code 0; tags are the
// tags the release created (the plan, when it ended well); f says whether the
// tags went to a remote.
// Returns: the title, the sentence and the detail lines.
func Result(ok bool, tags []string, f Facts) (title, why string, lines []string) {
	if !ok {
		return words.Say("tui.release_result.stopped"),
			words.Say("tui.release_result.why_stopped", "shortcutOf", commands.ShortcutOf("view.projects")),
			[]string{words.Say("tui.release_result.next_failed")}
	}
	lines = append(lines, words.Say("tui.release_result.tags"))
	for _, t := range tags {
		lines = append(lines, words.Say("tui.release_result.tag", "tag", t))
	}
	next := words.Say("tui.release_result.next_pushed")
	if f.NoPush || !f.HasRemote {
		next = words.Say("tui.release_result.next_local")
	}
	return words.Say("tui.release_result.ok"), words.Say("tui.release_result.why_ok"), append(lines, "", next)
}
