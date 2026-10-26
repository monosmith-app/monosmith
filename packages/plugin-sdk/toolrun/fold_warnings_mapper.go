package toolrun

import (
	"fmt"
	"regexp"
)

var (
	engineProblem = regexp.MustCompile(`(?m)^npm warn EBADENGINE\s+package: '([^']+)'`)
	engineHeader  = regexp.MustCompile(`(?m)^npm warn EBADENGINE\s+Unsupported engine`)
	deprecated    = regexp.MustCompile(`(?m)^npm warn deprecated\s`)
)

// FoldWarnings turns a package manager's warnings into one line each: two
// packages that want a newer Node become "2 packages want a newer Node", not
// two blocks of engine objects. It returns nothing when the output holds no
// warning it knows. detail names how to see the full output.
func FoldWarnings(output, detail string) []string {
	var notes []string
	if n := engineCount(output); n > 0 {
		notes = append(notes, fmt.Sprintf("%s newer Node than this one (%s)", plural(n, "package wants a", "packages want a"), detail))
	}
	if n := len(deprecated.FindAllString(output, -1)); n > 0 {
		notes = append(notes, fmt.Sprintf("%s (%s)", plural(n, "package is deprecated", "packages are deprecated"), detail))
	}
	return notes
}

func engineCount(output string) int {
	seen := map[string]bool{}
	for _, m := range engineProblem.FindAllStringSubmatch(output, -1) {
		seen[m[1]] = true
	}
	if len(seen) > 0 {
		return len(seen)
	}
	return len(engineHeader.FindAllString(output, -1))
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
