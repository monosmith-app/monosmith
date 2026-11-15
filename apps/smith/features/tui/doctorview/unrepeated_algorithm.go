package doctorview

import (
	"slices"
	"strings"
)

// Unrepeated drops from items every line the panel already said: a line that
// is, or sits inside, a sentence of said, a line that came twice, and a file
// line ("  writes <path>", written with two leading spaces) whose path is in
// said. A path then shows once.
//
// Parameters: items are the lines of "How to fix"; said are the lines above
// them, such as the problem.
// Returns: the lines that add something, in their order.
func Unrepeated(items, said []string) []string {
	seen := make([]string, 0, len(said)+len(items))
	for _, s := range said {
		seen = append(seen, norm(s))
	}
	var out []string
	for _, it := range items {
		n := norm(it)
		if n == "" || repeated(it, n, seen) {
			continue
		}
		seen = append(seen, n)
		out = append(out, it)
	}
	return out
}

// repeated is true when n is in one of the lines seen, or when item is a file
// line whose path is.
func repeated(item, n string, seen []string) bool {
	path := ""
	if strings.HasPrefix(item, "  ") {
		fields := strings.Fields(item)
		if len(fields) > 0 && strings.ContainsAny(fields[len(fields)-1], `/\`) {
			path = strings.ToLower(fields[len(fields)-1])
		}
	}
	for _, s := range seen {
		if strings.Contains(s, n) || (path != "" && strings.Contains(s, path)) {
			return true
		}
	}
	return false
}

// norm is a line in lower case with its spaces collapsed.
func norm(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

// TellOnce tells a check so that each text and each path shows once. The
// problem is cut into its sentences, with the build metadata of the tools
// moved to the notes; the lines of "How to fix" lose their own metadata, what
// the problem said already, and a path the problem already named (the file
// name stands for it).
//
// Parameters: check is the check's title; detail is its problem as a plugin
// wrote it; how are the lines of "How to fix".
// Returns: problems (one line each), notes (the build metadata, to fold away)
// and the lines of "How to fix" that remain.
func TellOnce(check, detail string, how []string) (problems, notes, items []string) {
	for _, part := range SplitProblem(detail) {
		cause, note := SplitBuildNotes(part)
		problems = appendSet(problems, cause)
		notes = appendSet(notes, note)
	}
	var lines []string
	for _, it := range how {
		cause, note := SplitBuildNotes(it)
		notes = appendSet(notes, note)
		if strings.HasPrefix(it, "  ") && cause != "" {
			cause = "  " + cause
		}
		lines = append(lines, cause)
	}
	items = Unrepeated(lines, append(append([]string{check}, problems...), notes...))
	for i, it := range items {
		items[i] = shortenPaths(it, problems)
	}
	return problems, notes, items
}

// appendSet adds text to list unless it is empty or already there.
func appendSet(list []string, text string) []string {
	if text == "" || slices.Contains(list, text) {
		return list
	}
	return append(list, text)
}

// shortenPaths replaces each path of line that a sentence of said already
// names by the name of its file.
func shortenPaths(line string, said []string) string {
	for _, field := range strings.Fields(line) {
		path := strings.TrimRight(field, ".,;:)")
		if len(path) < 8 || !strings.ContainsAny(path, `/\`) {
			continue
		}
		for _, s := range said {
			if strings.Contains(s, path) {
				line = strings.Replace(line, path, path[strings.LastIndexAny(path, `/\`)+1:], 1)
				break
			}
		}
	}
	return line
}
