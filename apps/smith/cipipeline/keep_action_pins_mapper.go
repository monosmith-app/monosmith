package cipipeline

import (
	"regexp"
	"strconv"
	"strings"
)

var usesPattern = regexp.MustCompile(`(uses:[ \t]+)([\w./-]+)@(v\d+(?:\.\d+){0,2})\b`)

// KeepNewerPins keeps, in the newly rendered file, the version of each action
// the old file pins higher than the template does. A dependency bot bumps the
// actions in a generated workflow, and the templates are code it cannot see:
// without this, every regeneration undoes its bumps and the doctor calls the
// file out of date. A pin the old file has lower, or in another form (a
// branch, a commit), is not kept, so a template's upgrade still reaches it.
func KeepNewerPins(old, rendered string) string {
	pinned := map[string]string{}
	for _, m := range usesPattern.FindAllStringSubmatch(old, -1) {
		if cur, ok := pinned[m[2]]; !ok || newerVersion(m[3], cur) {
			pinned[m[2]] = m[3]
		}
	}
	return usesPattern.ReplaceAllStringFunc(rendered, func(use string) string {
		m := usesPattern.FindStringSubmatch(use)
		if kept, ok := pinned[m[2]]; ok && newerVersion(kept, m[3]) {
			return m[1] + m[2] + "@" + kept
		}
		return use
	})
}

// Preserve is what a regeneration keeps of the file it replaces: the team's
// steps in the smith:slot markers and the action versions pinned newer.
func Preserve(old, rendered string) (content string, lost []string) {
	content, lost = KeepSlots(old, rendered)
	return KeepNewerPins(old, content), lost
}

// newerVersion reports whether a ("v6", "v1.2") is a higher version than b.
func newerVersion(a, b string) bool {
	as, bs := versionParts(a), versionParts(b)
	for i := 0; i < len(as) || i < len(bs); i++ {
		x, y := part(as, i), part(bs, i)
		if x != y {
			return x > y
		}
	}
	return false
}

func versionParts(v string) []string { return strings.Split(strings.TrimPrefix(v, "v"), ".") }

func part(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}
