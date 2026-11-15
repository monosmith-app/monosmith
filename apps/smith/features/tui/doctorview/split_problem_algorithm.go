package doctorview

import "strings"

// SplitProblem breaks the problem of a check into its sentences. A check that
// found many things joins them with "; " (one sentence per project), which
// read as a single paragraph fills a pane; as a list each one can be read,
// counted and folded. An empty problem gives no items.
func SplitProblem(problem string) []string {
	var out []string
	for _, part := range strings.Split(problem, "; ") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
