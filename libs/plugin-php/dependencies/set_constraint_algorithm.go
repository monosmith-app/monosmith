package dependencies

import (
	"regexp"
	"strings"
)

// SetConstraint sets the version constraint of one package in a
// composer.json source, in the section "require" or "require-dev", and says
// whether it found the package. Only that string changes: the file's
// indentation, key order and everything else stay as written.
func SetConstraint(src, section, name, constraint string) (string, bool) {
	open := regexp.MustCompile(`"` + regexp.QuoteMeta(section) + `"\s*:\s*\{`).FindStringIndex(src)
	if open == nil {
		return src, false
	}
	end := strings.Index(src[open[1]:], "}")
	if end < 0 {
		return src, false
	}
	body := src[open[1] : open[1]+end]
	entry := regexp.MustCompile(`("` + regexp.QuoteMeta(name) + `"\s*:\s*")[^"]*(")`)
	if !entry.MatchString(body) {
		return src, false
	}
	changed := entry.ReplaceAllString(body, "${1}"+constraint+"${2}")
	return src[:open[1]] + changed + src[open[1]+end:], true
}
