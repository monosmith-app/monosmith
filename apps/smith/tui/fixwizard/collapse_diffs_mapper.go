package fixwizard

import (
	"fmt"
	"strconv"
	"strings"
)

// summarizeDiffs replaces each unified diff in a review text by one line, so a
// fix that touches a few files opens on a short list and not on a hundred
// lines of change. The rest of the text stays as it is. It reports whether it
// found a diff, so the dialog knows whether a "show the changes" key is worth
// offering.
//
// A hunk header ("@@ -a,b +c,d @@") says how many lines the hunk holds, so a
// context line that looks like a step line cannot end a diff too early.
func summarizeDiffs(lines []string) (out []string, found bool) {
	for i := 0; i < len(lines); {
		path, ok := diffPath(lines, i)
		if !ok {
			out = append(out, lines[i])
			i++
			continue
		}
		found = true
		i += 2
		added, removed, before, after := 0, 0, 0, 0
		for i < len(lines) && strings.HasPrefix(lines[i], "@@ ") {
			aLen, bLen := hunkLengths(lines[i])
			before += aLen
			after += bLen
			i++
			for aRem, bRem := aLen, bLen; (aRem > 0 || bRem > 0) && i < len(lines); i++ {
				switch lines[i][:min(1, len(lines[i]))] {
				case "-":
					removed++
					aRem--
				case "+":
					added++
					bRem--
				default:
					aRem--
					bRem--
				}
			}
		}
		out = append(out, summaryLine(path, added, removed, before, after))
	}
	return out, found
}

// diffPath reports the file a diff at lines[i] is about.
func diffPath(lines []string, i int) (string, bool) {
	if i+1 >= len(lines) || !strings.HasPrefix(lines[i], "--- a/") || !strings.HasPrefix(lines[i+1], "+++ b/") {
		return "", false
	}
	return strings.TrimRight(strings.TrimPrefix(lines[i], "--- a/"), "\r"), true
}

// hunkLengths reads the two line counts of a hunk header; a missing count is 1.
func hunkLengths(header string) (before, after int) {
	for _, f := range strings.Fields(header) {
		if len(f) < 2 || (f[0] != '-' && f[0] != '+') {
			continue
		}
		n := 1
		if _, count, ok := strings.Cut(f[1:], ","); ok {
			n, _ = strconv.Atoi(count)
		}
		if f[0] == '-' {
			before = n
		} else {
			after = n
		}
	}
	return before, after
}

// summaryLine is the one line for a file: what happened and how much.
func summaryLine(path string, added, removed, before, after int) string {
	what := "changed"
	switch {
	case before == 0:
		what = "new file"
	case after == 0:
		what = "removed"
	}
	return fmt.Sprintf("  %s  %s  +%d -%d", path, what, added, removed)
}
