package operation

import "errors"

// VerifyFailure marks the error of an operation that did its work and then
// failed its own check (the new project exists, a tool it needs is missing).
// Wrap keeps the operation's record for such a run, so smith commit still
// knows its files, and settles the commit the person asked for.
type VerifyFailure struct {
	// Err is the failure the person reads.
	Err error
}

// Error is the text of the wrapped failure.
func (e *VerifyFailure) Error() string { return e.Err.Error() }

// Unwrap is the wrapped failure.
func (e *VerifyFailure) Unwrap() error { return e.Err }

// AfterWork marks err as a failed check after the operation changed files.
//
// Parameters: err is the failure of the check; nil stays nil.
// Returns: err wrapped as a VerifyFailure, or nil.
func AfterWork(err error) error {
	if err == nil {
		return nil
	}
	return &VerifyFailure{Err: err}
}

// isVerifyFailure is true when err is, or wraps, a VerifyFailure.
func isVerifyFailure(err error) bool {
	var v *VerifyFailure
	return errors.As(err, &v)
}
