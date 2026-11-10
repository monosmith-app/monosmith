package pathfinder

import "strings"

// splitValue cuts typed text at its last slash: dir is the folder part, with
// its closing slash ("" for the folder smith started in), and prefix is the
// name being typed. A backslash counts as a slash.
func splitValue(value string) (dir, prefix string) {
	v := strings.ReplaceAll(value, `\`, "/")
	i := strings.LastIndex(v, "/")
	return v[:i+1], v[i+1:]
}

// parentOf is the folder above dir ("a/b/" gives "a/", "a/" gives "", "" gives
// "../"). A folder that already climbs ("../") climbs one more.
func parentOf(dir string) string {
	trimmed := strings.TrimSuffix(dir, "/")
	if trimmed == "" {
		return "../"
	}
	i := strings.LastIndex(trimmed, "/")
	if trimmed[i+1:] == ".." {
		return dir + "../"
	}
	if i < 0 {
		return ""
	}
	return trimmed[:i+1]
}
