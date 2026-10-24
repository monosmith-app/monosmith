package fixwizard

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// reviewLines paints the text doctorfix.ShowFix wrote: added lines green,
// removed lines red, hunk headers muted, the fix's own line bold. Long lines
// wrap under themselves so nothing of a diff is cut off.
func reviewLines(th theme.Theme, text string, w int) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		paint := func(s string) string { return s }
		switch {
		case strings.HasPrefix(line, "●"):
			paint = th.Bold
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			paint = th.Bold
		case strings.HasPrefix(line, "@@"):
			paint = th.Muted
		case strings.HasPrefix(line, "+"):
			paint = th.Success
		case strings.HasPrefix(line, "-"):
			paint = th.Failure
		}
		for _, part := range pane.Wrap(line, w, 2) {
			out = append(out, paint(part))
		}
	}
	return out
}
