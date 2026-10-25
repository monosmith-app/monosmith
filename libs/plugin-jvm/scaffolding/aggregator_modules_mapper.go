package scaffolding

import (
	"regexp"
	"strings"
)

var modulesEnd = regexp.MustCompile(`(?m)^([ \t]*)</modules>`)

// AddModule lists dir in the aggregator pom's <modules>, adding the list
// before <properties> (or </project>) when there is none.
func AddModule(src, dir string) (string, bool) {
	entry := "<module>" + dir + "</module>"
	if strings.Contains(src, entry) {
		return src, false
	}
	if loc := modulesEnd.FindStringSubmatchIndex(src); loc != nil {
		indent := src[loc[2]:loc[3]]
		return src[:loc[0]] + indent + "  " + entry + "\n" + src[loc[0]:], true
	}
	block := "  <modules>\n    " + entry + "\n  </modules>\n\n"
	for _, anchor := range []string{"  <properties>", "</project>"} {
		if i := strings.Index(src, anchor); i >= 0 {
			return src[:i] + block + src[i:], true
		}
	}
	return src, false
}

// RemoveModule takes dir out of the aggregator pom's <modules>, with its line.
func RemoveModule(src, dir string) (string, bool) {
	line := regexp.MustCompile(`(?m)^[ \t]*<module>\s*` + regexp.QuoteMeta(dir) + `\s*</module>[ \t]*\r?\n?`)
	out := line.ReplaceAllString(src, "")
	return out, out != src
}
