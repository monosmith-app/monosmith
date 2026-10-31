package lineendings

import (
	"path"
	"strings"
)

// EOLFromAttributes reads the text of a git attributes file and returns the
// line ending it sets for rel, a slash-separated path from the file's folder:
// "crlf", "lf", or "" when no rule sets one. The last matching rule wins. A
// file marked binary or -text has no line ending to set: it returns "".
func EOLFromAttributes(attributes, rel string) string {
	eol := ""
	for _, line := range strings.Split(attributes, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") || !matchesPattern(fields[0], rel) {
			continue
		}
		for _, attr := range fields[1:] {
			switch attr {
			case "eol=crlf":
				eol = "crlf"
			case "eol=lf":
				eol = "lf"
			case "binary", "-text", "-eol":
				eol = ""
			}
		}
	}
	return eol
}

// matchesPattern applies the common attribute patterns: a name or glob
// without a slash matches the file's name at any depth, one with a slash
// matches the path from the root.
func matchesPattern(pattern, rel string) bool {
	pattern = strings.TrimPrefix(pattern, "/")
	if !strings.Contains(pattern, "/") {
		ok, _ := path.Match(pattern, path.Base(rel))
		return ok
	}
	if rest, found := strings.CutPrefix(pattern, "**/"); found {
		// "**/" stands for any number of leading folders: try each tail of rel.
		for p := rel; ; {
			if ok, _ := path.Match(rest, p); ok {
				return true
			}
			_, tail, more := strings.Cut(p, "/")
			if !more {
				return false
			}
			p = tail
		}
	}
	ok, _ := path.Match(pattern, rel)
	return ok
}
