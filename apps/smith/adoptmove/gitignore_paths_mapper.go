package adoptmove

import "strings"

// FixGitignore rewrites the lines of a .gitignore that name a moved top-level
// entry from the repository root (a leading slash: /dist, /coverage/, !/keep)
// so they point at the entry's new place under dir. Lines without a leading
// slash match at any depth and need no change. moved are the top-level names
// that moved. It returns the new text and whether anything changed.
func FixGitignore(text, dir string, moved []string) (string, bool) {
	isMoved := map[string]bool{}
	for _, m := range moved {
		isMoved[m] = true
	}
	lines := strings.Split(text, "\n")
	changed := false
	for i, line := range lines {
		negate := ""
		body := strings.TrimRight(line, "\r")
		if strings.HasPrefix(body, "!") {
			negate, body = "!", body[1:]
		}
		if !strings.HasPrefix(body, "/") {
			continue
		}
		first, _, _ := strings.Cut(strings.TrimPrefix(body, "/"), "/")
		first = strings.TrimRight(first, " ")
		if first == "" || !isMoved[first] {
			continue
		}
		lines[i] = negate + "/" + dir + body
		if strings.HasSuffix(line, "\r") {
			lines[i] += "\r"
		}
		changed = true
	}
	return strings.Join(lines, "\n"), changed
}
