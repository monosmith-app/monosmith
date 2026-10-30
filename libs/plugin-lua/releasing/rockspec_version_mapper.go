package releasing

import "regexp"

var (
	versionField = regexp.MustCompile(`(?m)^(version\s*=\s*)(["'])[^"'\r\n]*(["'])`)
	packageLine  = regexp.MustCompile(`(?m)^package\s*=[^\n]*\n`)
)

// SetVersion writes version as the rockspec's version, with LuaRocks'
// revision reset to 1 (a new version restarts the revision), adding the field
// under package when there is none. A rockspec with neither is left alone.
func SetVersion(src, version string) (string, bool) {
	if loc := versionField.FindStringSubmatchIndex(src); loc != nil {
		out := versionField.ReplaceAllString(src, "${1}${2}"+version+"-1${3}")
		return out, out != src
	}
	if loc := packageLine.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + "version = \"" + version + "-1\"\n" + src[loc[1]:], true
	}
	return src, false
}
