package adoptsplit

import "strings"

// KeepsOwnVersion reports whether a development dependency stays in the
// project instead of going to the workspace, because the workspace has it at
// another major version. The root keeps one range for every project, so
// hoisting typescript ^5.9.3 beside a root typescript ^4.9.5 would give the
// project TypeScript 4 without a word.
//
// Parameters: rootRange is the range in the root package.json ("" when the
// root does not have the dependency); projectRange is the project's range.
// Returns: true when both ranges name a major version and the majors differ.
// A range without a plain major (a tag, a URL, "*") never keeps.
func KeepsOwnVersion(rootRange, projectRange string) bool {
	a, okA := rangeMajor(rootRange)
	b, okB := rangeMajor(projectRange)
	return okA && okB && a != b
}

// rangeMajor reads the major version at the start of an npm range such as
// ^5.9.3, ~4.1, >=18, v2 or 3.x. It returns false when the range does not
// start with one.
func rangeMajor(r string) (string, bool) {
	r = strings.TrimLeft(strings.TrimSpace(r), "^~>=<v ")
	end := 0
	for end < len(r) && r[end] >= '0' && r[end] <= '9' {
		end++
	}
	if end == 0 {
		return "", false
	}
	return r[:end], true
}
