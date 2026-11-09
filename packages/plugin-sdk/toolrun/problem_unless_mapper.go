package toolrun

// ProblemUnless is the Problem of a doctor finding: problem when the check
// failed, "" when it passed (ok is true), because a passing finding has none.
//
// Parameters: ok says whether the check passed; problem is the one-line
// statement of what was found wrong.
// Returns: problem, or "" for a passing check.
func ProblemUnless(ok bool, problem string) string {
	if ok {
		return ""
	}
	return problem
}
