package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// A diff of a CRLF file or of one indented with tabs must not move the cursor
// when it is drawn: every line has to be exactly what its width says.
func TestReviewLinesOfACRLFAndTabIndentedDiffFitTheDialog(t *testing.T) {
	text := "--- a/eslint.config.mjs\r\n+++ b/eslint.config.mjs\r\n@@ -1,4 +1,4 @@\r\n" +
		"-\tname: 'mnci/base',\r\n+\tignores: ['dist', 'coverage', '.venv', 'a very long entry that has to wrap'],\r\n"
	const w = 30
	for _, line := range reviewLines(plain(), text, w) {
		if strings.ContainsAny(line, "\r\t") {
			t.Fatalf("line %q still has a control character", line)
		}
		if pane.Width(line) > w {
			t.Fatalf("line %q is %d cells, wider than %d", line, pane.Width(line), w)
		}
	}
}
