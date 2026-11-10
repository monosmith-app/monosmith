package releasing

import (
	"regexp"
	"strings"
)

var (
	projectCall = regexp.MustCompile(`(?is)\bproject\s*\(\s*[A-Za-z0-9_.+-]+[^)]*\)`)
	versionArg  = regexp.MustCompile(`(?i)(\bVERSION\s+)[0-9][0-9.]*`)
	projectName = regexp.MustCompile(`(?is)^project\s*\(\s*[A-Za-z0-9_.+-]+`)
)

// SetVersion writes version into the project() call's VERSION, adding the
// argument after the project's name when there is none. CMake accepts only
// numbers and dots there, so a prerelease or build suffix (2.0.0-beta.1) is
// left out: the release tag carries it.
func SetVersion(src, version string) (string, bool) {
	if i := strings.IndexAny(version, "-+"); i >= 0 {
		version = version[:i]
	}
	loc := projectCall.FindStringIndex(src)
	if loc == nil {
		return src, false
	}
	call := src[loc[0]:loc[1]]
	var updated string
	if versionArg.MatchString(call) {
		updated = versionArg.ReplaceAllString(call, "${1}"+version)
	} else {
		name := projectName.FindStringIndex(call)
		updated = call[:name[1]] + " VERSION " + version + call[name[1]:]
	}
	out := src[:loc[0]] + updated + src[loc[1]:]
	return out, out != src
}
