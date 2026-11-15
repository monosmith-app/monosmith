package workspace

import (
	"regexp"
	"strings"
)

var tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([^\]]+?)\s*\]\]?\s*(#.*)?$`)

// excludeArray is a one-line `exclude = [...]` array, the list of projects a
// convention leaves alone.
var excludeArray = regexp.MustCompile(`^(\s*exclude\s*=\s*\[)([^\]]*)(\].*)$`)

// RemoveProjectTables deletes [projects.<name>] and every table under it
// ([projects.<name>.tasks.build], ...) from smith.toml text, with the comment
// lines directly above each removed header, and takes the name out of the
// exclude lists of the conventions. Every other line stays exactly: the
// comment directly above the table after a removed one belongs to that table,
// and a comment paragraph set apart by blank lines stays too. It reports
// whether anything was removed.
//
// Parameters: src is the text of smith.toml; name is the project's name.
// Returns: the new text, and whether anything changed.
func RemoveProjectTables(src, name string) (string, bool) {
	lines := strings.SplitAfter(src, "\n")
	headers := map[int]string{}
	for i, line := range lines {
		if m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			headers[i] = m[1]
		}
	}
	// owner is the header a line belongs to: its own section, or the header it
	// is a comment directly above.
	owner := make([]int, len(lines))
	current := -1
	for i := range lines {
		if _, ok := headers[i]; ok {
			current = i
		}
		owner[i] = current
	}
	for h := range headers {
		for j := h - 1; j >= 0 && isComment(lines[j]); j-- {
			owner[j] = h
		}
	}
	var out []string
	removed, changed := false, false
	table := ""
	paragraph := false // inside a comment paragraph set apart by a blank line
	for i, line := range lines {
		if key, ok := headers[i]; ok {
			table = key
		}
		if h := owner[i]; h >= 0 && projectTable(headers[h], name) {
			switch {
			case strings.TrimSpace(line) == "":
				paragraph = false
				out = append(out, line)
			case i > h && isComment(line) && (paragraph || strings.TrimSpace(lines[i-1]) == ""):
				// A comment paragraph set apart by a blank line is nobody's to delete.
				paragraph = true
				out = append(out, line)
			default:
				paragraph = false
				removed = true
			}
			continue
		}
		paragraph = false
		if strings.HasPrefix(table, "conventions.") {
			if next, ok := dropFromExclude(line, name); ok {
				out, changed = append(out, next), true
				continue
			}
		}
		out = append(out, line)
	}
	text := strings.Join(out, "")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text, removed || changed
}

// recipeCommentOf returns the project a line's comment is about when it is
// the one `smith add` writes above the tables a recipe declares for a project
// ("# Written by smith from the recipe team for the project billing."), else "".
func recipeCommentOf(line string) string {
	rest, ok := strings.CutPrefix(strings.TrimRight(line, "\r\n"), "# Written by smith from the recipe ")
	if !ok {
		return ""
	}
	_, project, found := strings.Cut(rest, " for the project ")
	if !found {
		return ""
	}
	return strings.TrimSuffix(project, ".")
}

// isComment reports whether a line holds only a comment.
func isComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// dropFromExclude takes a project's name out of a one-line exclude array.
// It reports false when the line is not such an array or does not name it.
func dropFromExclude(line, name string) (string, bool) {
	m := excludeArray.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
	if m == nil {
		return line, false
	}
	var kept []string
	found := false
	for _, item := range strings.Split(m[2], ",") {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if strings.Trim(value, `"'`) == name {
			found = true
			continue
		}
		kept = append(kept, value)
	}
	if !found {
		return line, false
	}
	end := line[len(strings.TrimRight(line, "\r\n")):]
	return m[1] + strings.Join(kept, ", ") + m[3] + end, true
}

// projectTable reports whether a table key is projects.<name> or below it.
func projectTable(key, name string) bool {
	rest, ok := strings.CutPrefix(key, "projects.")
	if !ok {
		return false
	}
	for _, form := range []string{name, `"` + name + `"`} {
		if rest == form || strings.HasPrefix(rest, form+".") {
			return true
		}
	}
	return false
}
