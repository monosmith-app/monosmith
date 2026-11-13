package lintfix

// Outcome is what `smith lint --fix` found in one project: how many problems
// the fix removed, and what is left.
type Outcome struct {
	Project string
	// Fixed is the number of problems the fix removed; -1 when the tool does
	// not let smith count them.
	Fixed int
	// Counted says Errors and Warnings are what is left. When it is false the
	// tool only says whether problems remain: Left.
	Counted          bool
	Errors, Warnings int
	// Left says problems remain, for a tool that does not count them.
	Left bool
	// NoFix says the tool has no fix mode: the plain lint ran and Left is its
	// verdict.
	NoFix bool
	// Skipped says the lint did not run, because a task it needs failed.
	Skipped bool
	// Crash is the first line of the error of a tool that could not run, and
	// Remedy says what to do about it.
	Crash, Remedy string
}

// Problems returns the number of problems left, or 1 for a project whose tool
// says only that some remain.
func (o Outcome) Problems() int {
	if o.Counted {
		return o.Errors + o.Warnings
	}
	if o.Left {
		return 1
	}
	return 0
}
