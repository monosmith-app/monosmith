package doctor

import (
	"fmt"
	"strings"
)

// shownProblems is how many problems of one check a finding spells out before
// it says how many more there are.
const shownProblems = 3

// Summary joins the first shown items with sep and counts the rest, so a check
// that fails on 23 projects says so in a few lines.
func Summary(items []string, sep string, shown int) string {
	if len(items) <= shown {
		return strings.Join(items, sep)
	}
	more := fmt.Sprintf("and %d more", len(items)-shown)
	return strings.Join(items[:shown], sep) + sep + more
}
