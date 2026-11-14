package statuspaint

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// statusLine matches a status line of a command: an optional indent and task id,
// then one glyph and a space ("kit:build ✓ 106ms", "  × web:build reason",
// "✓ 2/2 tasks succeeded"). A line of a tool is "id │ text" and never matches.
var statusLine = regexp.MustCompile(`^(\s*(?:[^\s│]+ )?)(` + marks.Passed + `|` + marks.Failed + `|` + marks.Warning + `|` + marks.Skipped + `|▶|■|→|!|≈) `)

// statusCodes are the colours of the glyphs, from the log palette: green for
// passed, red for failed, yellow for a warning, dim for what did not run or only
// started or points on.
var statusCodes = map[string]string{
	marks.Passed: termcolour.Green, "≈": termcolour.Green,
	marks.Failed:  termcolour.Red,
	marks.Warning: termcolour.Yellow, "!": termcolour.Yellow,
	marks.Skipped: termcolour.Dim, "▶": termcolour.Dim, "■": termcolour.Dim, "→": termcolour.Dim,
}

// Paint colours the glyph of every status line of the text, and nothing
// else: the words and the numbers keep the colour of the terminal, and a glyph
// is never the only carrier of the meaning. A line that is not a status line, or
// that carries colour already, is left as it is.
//
// Parameters: text is one or more lines of a command's own report, not tool output.
// Returns: the text with its glyphs painted.
func Paint(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if termcolour.HasSGR(line) {
			continue
		}
		if m := statusLine.FindStringSubmatch(line); m != nil {
			lines[i] = m[1] + termcolour.Wrap(m[2], statusCodes[m[2]]) + line[len(m[1])+len(m[2]):]
		}
	}
	return strings.Join(lines, "\n")
}
