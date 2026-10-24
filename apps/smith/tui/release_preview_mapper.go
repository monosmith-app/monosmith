package tui

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// isPreview is true for the command line of the release preview.
func isPreview(argv []string) bool {
	return len(argv) > 0 && argv[0] == "release" && slices.Contains(argv, "--dry-run")
}

// previewLines is the Release screen's own view of the preview run: what it
// is waiting for while it runs, and its output when it has ended, so the
// preview does not need the Run screen. room is how many lines it may take.
func (s Shell) previewLines(w, room int) []string {
	if s.run == nil || !isPreview(s.run.Command) || room < 3 {
		return nil
	}
	lines := []string{"", s.th.Title("Preview: " + s.run.CommandLine())}
	if s.running() {
		wait := "Waiting for each package's registry (is this version already published?) and for the git history. " +
			"With many packages this takes a while. Nothing is changed. Ctrl+C stops it."
		lines = append(lines, pane.Wrap(s.th.Info(s.th.Spin(s.tick))+" "+wait, w, 2)...)
		return capTail(lines, room)
	}
	mark := s.th.Success(s.th.Icons.Passed) + " done"
	if !s.run.OK() {
		mark = s.th.Failure(s.th.Icons.Failed) + " failed"
	}
	lines = append(lines, mark+s.th.Muted(" · full output on the Run screen (Alt+2)"))
	var out []string
	for _, l := range s.run.Lines {
		out = append(out, pane.Wrap(pane.Strip(l), w-2, 2)...)
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

// capTail keeps the first n lines.
func capTail(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}
