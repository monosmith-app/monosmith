package doctor

import "fmt"

// foldAbove is how many passed lines doctor lists before it folds them into a
// count: the useful part of a long report is what failed and what to do.
const foldAbove = 8

// FoldPassed replaces a long list of passed lines by one line with their count.
//
// Parameters: lines are the passed lines after CollapsePassed.
// Returns: lines unchanged when there are foldAbove or fewer; otherwise one line,
// "N checks passed (--verbose lists them)".
func FoldPassed(lines []PassedLine) []PassedLine {
	if len(lines) <= foldAbove {
		return lines
	}
	total := 0
	for _, l := range lines {
		total += l.Count
	}
	return []PassedLine{{Text: fmt.Sprintf("%d checks passed (--verbose lists them)", total), Count: total}}
}
