package adoptmove

import "strings"

// FixGitignore rewrites the lines of a .gitignore that name a top-level entry
// from the repository root (a leading slash: /dist, /coverage/, !/keep) for a
// project that moved into dir. Lines without a leading slash match at any depth
// and need no change.
//
// A line that names a moved entry now points at its new place under dir. A
// line that names an entry the repository does not track (build output such as
// /dist or /node_modules, which the project makes in its new folder from now
// on) keeps its root line and gets a second line under dir. A line that names
// an entry that stays at the root (in stays) does not change.
//
// text is the .gitignore; dir is the project's folder; moved are the top-level
// names that moved; stays are the tracked top-level names that stay at the
// root. It returns the new text and whether anything changed.
func FixGitignore(text, dir string, moved, stays []string) (string, bool) {
	isMoved := map[string]bool{}
	for _, m := range moved {
		isMoved[m] = true
	}
	stay := map[string]bool{}
	for _, s := range stays {
		stay[s] = true
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		out = append(out, line)
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
		if first == "" || stay[first] || strings.ContainsAny(first, "*?[") {
			continue
		}
		moving := negate + "/" + dir + body
		if strings.HasSuffix(line, "\r") {
			moving += "\r"
		}
		if isMoved[first] {
			out[len(out)-1] = moving
		} else {
			out = append(out, moving)
		}
		changed = true
	}
	return strings.Join(out, "\n"), changed
}
