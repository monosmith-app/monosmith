package commandtext

import (
	"slices"
	"strings"
)

// SaidNotCommitted is true when a command asked to commit but printed that it
// did not (for example because the work tree had other changes): there is then
// nothing for smith undo to revert.
//
// Parameters: lines are the lines the command printed.
// Returns: whether one of them starts with "Not committed".
func SaidNotCommitted(lines []string) bool {
	return slices.ContainsFunc(lines, func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "Not committed") }) // words:literal
}
