package tui

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// isPreview is true for the command line of the release preview.
func isPreview(argv []string) bool {
	return len(argv) > 0 && argv[0] == "release" && slices.Contains(argv, "--dry-run")
}

// previewLines is the Release screen's own view of the preview run: what it
// is waiting for while it runs, and its output when it has ended, so the
// preview does not need the run panel. room is how many lines it may take.
func (s Shell) previewLines(w, room int) []string {
	if s.run != nil && isRelease(s.run.Command) {
		return s.releaseRunLines(w, room)
	}
	if s.run == nil || !isPreview(s.run.Command) || room < 3 {
		return nil
	}
	lines := []string{"", s.th.Title(words.Say("tui.release_preview.preview", "commandLine", s.run.CommandLine()))}
	if s.running() {
		wait := words.Say("tui.release_preview.waiting_for_each_package_s_registry")
		lines = append(lines, pane.Wrap(s.th.Info(s.th.Spin(s.tick))+" "+wait, w, 2)...)
		return capTail(lines, room)
	}
	mark := words.Say("tui.release_preview.done", "passed", s.th.Success(s.th.Icons.Passed))
	if !s.run.OK() {
		mark = words.Say("tui.release_preview.failed", "failed", s.th.Failure(s.th.Icons.Failed))
	}
	lines = append(lines, mark+s.th.Muted(words.Say("tui.release_preview.full_output_in_the_run_panel", "shortcutOf", commands.ShortcutOf("view.projects"))))
	var out []string
	for _, l := range s.previewOutput() {
		out = append(out, pane.Wrap(pane.Strip(l), w-2, 2)...)
	}
	if s.release.CanRelease() {
		out = append(out, s.th.Muted(words.Say("tui.release_preview.next", "key", commands.ShortcutOf("rel.release"))))
	}
	keep := max(room-len(lines), 1)
	if len(out) > keep {
		out = append([]string{s.th.Muted("…")}, out[len(out)-(keep-1):]...)
	}
	for _, l := range out {
		lines = append(lines, "  "+l)
	}
	return lines
}

// previewOutput is what the preview printed, less what the view says itself: an
// empty plan is the headline of the view (said once, not twice), and the closing
// line that sends the person to a terminal is the button of the view.
func (s Shell) previewOutput() []string {
	if s.release.Empty() {
		return nil
	}
	var out []string
	for _, l := range s.run.Lines {
		if strings.TrimSpace(pane.Strip(l)) != release.PublishHintLine {
			out = append(out, l)
		}
	}
	return out
}

// capTail keeps the first n lines.
func capTail(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}
