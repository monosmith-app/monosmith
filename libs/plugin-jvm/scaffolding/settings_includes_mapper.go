package scaffolding

import (
	"regexp"
	"strings"
)

var (
	includeLine = regexp.MustCompile(`(?m)^include\(.*\)[ \t]*$`)
	rootName    = regexp.MustCompile(`(?m)^rootProject\.name\s*=\s*["']([^"']+)["']`)
)

// AddInclude includes the project at path (":apps:api") in the settings,
// after the last include, or at the end when there is none.
func AddInclude(src, path string) (string, bool) {
	line := `include("` + path + `")`
	if strings.Contains(src, line) {
		return src, false
	}
	if locs := includeLine.FindAllStringIndex(src, -1); locs != nil {
		at := locs[len(locs)-1][1]
		return src[:at] + "\n" + line + src[at:], true
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += "\n"
	}
	return src + "\n" + line + "\n", true
}

// RootProjectName is the settings' rootProject.name, or empty.
func RootProjectName(src string) string {
	if m := rootName.FindStringSubmatch(src); m != nil {
		return m[1]
	}
	return ""
}

// RemoveInclude takes the include of the project at path (":apps:api") out of
// the settings, with its line.
func RemoveInclude(src, path string) (string, bool) {
	line := regexp.MustCompile(`(?m)^include\(\s*["']` + regexp.QuoteMeta(path) + `["']\s*\)[ \t]*\r?\n?`)
	out := line.ReplaceAllString(src, "")
	return out, out != src
}
