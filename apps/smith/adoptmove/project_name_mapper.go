package adoptmove

import (
	"regexp"
	"strings"
)

var (
	camelBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	unsafeRun     = regexp.MustCompile(`[^a-z0-9]+`)
	majorVersion  = regexp.MustCompile(`^v[0-9]+$`)
)

// DeriveName turns a manifest name into a project folder name: the npm scope
// or the module path in front is dropped, camelCase and separators become
// single hyphens, and the result is lowercase kebab-case that starts with a
// letter. fallback is tried when nothing usable is left (it goes through the
// same rules), and "project" is the last resort.
func DeriveName(raw, fallback string) string {
	if name := kebab(lastSegment(raw)); name != "" {
		return name
	}
	if name := kebab(lastSegment(fallback)); name != "" {
		return name
	}
	return "project"
}

// lastSegment is the part after the scope (@scope/name) or after the last slash
// of a module path; a trailing major version (/v2) is skipped.
func lastSegment(raw string) string {
	parts := strings.Split(strings.Trim(strings.TrimSpace(raw), "/"), "/")
	for len(parts) > 1 && majorVersion.MatchString(parts[len(parts)-1]) {
		parts = parts[:len(parts)-1]
	}
	return parts[len(parts)-1]
}

func kebab(s string) string {
	s = camelBoundary.ReplaceAllString(s, "${1}-${2}")
	s = strings.Trim(unsafeRun.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return ""
	}
	if len(s) > 50 {
		s = strings.TrimRight(s[:50], "-")
	}
	return s
}
