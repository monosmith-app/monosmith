package doctor

import "strconv"

// problemWhen is the Problem of a finding: text when the check failed, else
// "" because a passing finding names no problem.
//
// Parameters: failed is true when the finding is not OK; text is the one-line
// problem. Returns: text or "".
func problemWhen(failed bool, text string) string {
	if !failed {
		return ""
	}
	return text
}

// leadProblem words a list of problems as one line: the first, and how many
// more follow. It returns "" for an empty list.
//
// Parameters: items are the problems found, each a short sentence.
// Returns: the first item, plus " (and N more)" when there are others.
func leadProblem(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return items[0] + " (and " + strconv.Itoa(len(items)-1) + " more)"
}
