package releasing

import (
	"regexp"
	"strings"
)

var (
	// tableStart is the first table of a recipe file: only what comes before
	// it is the recipe's own top-level keys.
	tableStart  = regexp.MustCompile(`(?m)^[ \t]*\[`)
	versionLine = regexp.MustCompile(`(?m)^version[ \t]*=[ \t]*(?:"[^"\r\n]*"|'[^'\r\n]*')`)
	nameLine    = regexp.MustCompile(`(?m)^name[ \t]*=[^\r\n]*`)
)

// SetVersion writes version as the recipe's top-level version, keeping the
// rest of the file as it is: the line is replaced where it is, else added
// after the name, else at the top. It reports whether the file changed.
func SetVersion(src, version string) (string, bool) {
	head, tail := src, ""
	if loc := tableStart.FindStringIndex(src); loc != nil {
		head, tail = src[:loc[0]], src[loc[0]:]
	}
	line := `version = "` + version + `"`
	switch {
	case versionLine.MatchString(head):
		loc := versionLine.FindStringIndex(head)
		head = head[:loc[0]] + line + head[loc[1]:]
	case nameLine.MatchString(head):
		loc := nameLine.FindStringIndex(head)
		newline := "\n"
		if strings.Contains(head, "\r\n") {
			newline = "\r\n"
		}
		head = head[:loc[1]] + newline + line + head[loc[1]:]
	default:
		head = line + "\n" + head
	}
	out := head + tail
	return out, out != src
}
