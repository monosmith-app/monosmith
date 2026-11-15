package pluginpack

import (
	"regexp"
	"strings"
)

// programPattern turns the first element of a manifest command into a matcher
// for the files that can stand for it on some platform: {os} and {arch} are
// any word, {exe} is ".exe" or nothing. ok is false when the element does not
// name a file of the plugin (a program found on PATH).
func programPattern(first string) (re *regexp.Regexp, ok bool) {
	if !strings.HasPrefix(first, "./") && !strings.HasPrefix(first, "../") {
		return nil, false
	}
	if strings.HasPrefix(first, "../") {
		return nil, false
	}
	rest := strings.TrimPrefix(first, "./")
	var b strings.Builder
	b.WriteString("^")
	for rest != "" {
		switch {
		case strings.HasPrefix(rest, "{os}"), strings.HasPrefix(rest, "{arch}"):
			b.WriteString("[a-z0-9_]+")
			rest = rest[strings.IndexByte(rest, '}')+1:]
		case strings.HasPrefix(rest, "{exe}"):
			b.WriteString(`(\.exe)?`)
			rest = rest[len("{exe}"):]
		default:
			next := strings.IndexByte(rest, '{')
			if next < 0 {
				next = len(rest)
			}
			if next == 0 {
				next = 1
			}
			b.WriteString(regexp.QuoteMeta(rest[:next]))
			rest = rest[next:]
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String()), true
}
