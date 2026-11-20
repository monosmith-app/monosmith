package release

import "io"

// progressWriter picks where a release run prints its progress lines. With
// --json, stdout holds only the report, so a script can parse it, and the
// progress goes to stderr.
//
// Parameters: asJSON is the --json flag; out is stdout; errOut is stderr.
// Returns the writer for progress lines.
func progressWriter(asJSON bool, out, errOut io.Writer) io.Writer {
	if asJSON {
		return errOut
	}
	return out
}
