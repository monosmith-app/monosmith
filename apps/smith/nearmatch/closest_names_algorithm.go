package nearmatch

import (
	"sort"
	"strings"
)

// Closest returns the candidates a mistyped name was probably meant to be:
// those within two edits of it, or that contain it or are contained by it,
// nearest first. It is empty when nothing is close.
//
// Parameters: name is what the person typed; candidates are the valid names.
func Closest(name string, candidates []string) []string {
	name = strings.ToLower(name)
	type scored struct {
		name string
		d    int
	}
	var found []scored
	for _, c := range candidates {
		lc := strings.ToLower(c)
		d := distance(name, lc)
		if d <= 2 || (len(name) >= 3 && (strings.Contains(lc, name) || strings.Contains(name, lc))) {
			found = append(found, scored{c, d})
		}
	}
	sort.SliceStable(found, func(i, j int) bool { return found[i].d < found[j].d })
	out := make([]string, 0, len(found))
	for _, f := range found {
		out = append(out, f.name)
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}

// Suggestion is the sentence to append to an "unknown name" error: a space and
// "Did you mean a or b?", or "" when nothing is close.
func Suggestion(name string, candidates []string) string {
	near := Closest(name, candidates)
	if len(near) == 0 {
		return ""
	}
	return " Did you mean " + strings.Join(near, " or ") + "?"
}

// distance is the edit distance between two names.
func distance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
