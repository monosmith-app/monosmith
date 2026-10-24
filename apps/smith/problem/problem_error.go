package problem

import (
	"fmt"
	"strings"
)

// Error is a problem with the step that clears it. A message that ends in a
// remedy is one the person can act on without reading the docs.
type Error struct {
	// Problem is one short sentence naming what is wrong.
	Problem string
	// Remedy is what to run or do next; each entry is one option.
	Remedy []string
	// Cause is the sentinel or error this one stands for, for errors.Is.
	Cause error
}

// New is a problem and its remedy options.
func New(problem string, remedy ...string) *Error {
	return &Error{Problem: problem, Remedy: remedy}
}

// Error prints the problem on one line and each remedy on an arrow line.
func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "✖ %s", e.Problem)
	for _, r := range e.Remedy {
		fmt.Fprintf(&b, "\n  → %s", r)
	}
	return b.String()
}

// Unwrap is the error this one explains.
func (e *Error) Unwrap() error { return e.Cause }
