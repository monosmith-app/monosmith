package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/releaseconfirm"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// releaseRows are the lines under the plan of the Release view: the checks
// before a release in a box, the sentence that nothing is published unasked,
// the buttons Release now, Dry run and Release notes, and their command
// lines, dim. Release now asks before it runs. They are drawn only while the
// plan can be released and no release runs; while a preview shows, the checks
// leave their room to its output.
func (s Shell) releaseRows(w int) []buttonRow {
	if !s.release.CanRelease() || (s.run != nil && isRelease(s.run.Command)) || w < 24 {
		return nil
	}
	var out []buttonRow
	if s.run == nil || !isPreview(s.run.Command) {
		out = append(out, s.releaseChecks(w)...)
	}
	buttons := []viewButton{{"rel.release", words.Say("tui.release_confirm.run")}, {"rel.preview", words.Say("releasepage.button.preview")}, {"rel.notes", words.Say("releasepage.button.notes")}}
	out = append(out, buttonRow{})
	out = append(out, s.buttonLines(buttons, w-1, s.actionAt(s.mouseX, s.mouseY))...)
	for _, line := range pane.Wrap(s.th.Faint(words.Say("releasepage.commands")), w, 2) {
		out = append(out, buttonRow{line: line})
	}
	return out
}

// releaseChecks are the checks before a release in a box titled "Checks before
// release", and the sentence that nothing is published until the plan is
// confirmed.
func (s Shell) releaseChecks(w int) []buttonRow {
	var checks []string
	for _, c := range releaseconfirm.Preflight(s.releaseFacts(nil)) {
		mark := s.th.Success(s.th.Icons.Passed)
		if c.Warn {
			mark = s.th.Warning(s.th.Icons.Warning)
		}
		checks = append(checks, pane.Wrap(" "+mark+" "+c.Text, w-2, 3)...)
	}
	checks = append(checks, pane.Wrap(" "+s.th.Muted(s.th.Icons.Queued+" "+words.Say("releasepage.checks.registry")), w-2, 3)...)
	var out []buttonRow
	out = append(out, buttonRow{})
	for _, line := range pane.Frame(s.th, words.Say("releasepage.checks.title"), checks, w, len(checks)+2, false) {
		out = append(out, buttonRow{line: line})
	}
	return append(out, buttonRow{}, buttonRow{line: s.th.Muted(words.Say("releasepage.confirm_note"))})
}
