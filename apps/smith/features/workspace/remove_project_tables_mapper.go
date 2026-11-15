package workspace

import (
	"regexp"
	"strings"
)

var tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([^\]]+?)\s*\]\]?\s*(#.*)?$`)

// RemoveProjectTables deletes [projects.<name>] and every table under it
// ([projects.<name>.tasks.build], ...) from smith.toml text, keeping all other
// lines exactly. It reports whether anything was removed.
func RemoveProjectTables(src, name string) (string, bool) {
	lines := strings.SplitAfter(src, "\n")
	var out []string
	skipping, removed := false, false
	for _, line := range lines {
		if m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			starting := projectTable(m[1], name) && !skipping
			skipping = projectTable(m[1], name)
			removed = removed || skipping
			// The comment smith add wrote above the project's tables goes with them.
			if starting && len(out) > 0 && recipeCommentOf(out[len(out)-1]) == name {
				out = out[:len(out)-1]
			}
		} else if skipping && recipeCommentOf(line) != "" && recipeCommentOf(line) != name {
			// The comment above another project's tables sits after the end of
			// this one's: it is not this project's to delete.
			out = append(out, line)
			continue
		}
		if !skipping {
			out = append(out, line)
		}
	}
	text := strings.Join(out, "")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text, removed
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
