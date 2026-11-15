package workspace

import "errors"

// UpgradeCommand is the one command that installs the latest smith.
const UpgradeCommand = "npm install --global monosmith"

// NewerSmithError says smith.toml was written for a newer smith than the one
// reading it: its schema is newer, or it has tables or several keys this smith
// does not know. A single unknown key stays a plain error, because it is more
// likely a typo.
type NewerSmithError struct {
	// Detail is the full text of what this smith could not read: every parse
	// error, or the schema line. smith prints it only with --verbose.
	Detail string
}

// Error returns the detail, so callers that print the error as is keep the
// full text.
func (e *NewerSmithError) Error() string { return e.Detail }

// Summary is the one sentence smith prints for the error, for the running
// smith version.
func (e *NewerSmithError) Summary(version string) string {
	return "This smith (" + version + ") is older than smith.toml needs. Run `" + UpgradeCommand + "`."
}

// AsNewerSmith returns the NewerSmithError inside err, if there is one.
func AsNewerSmith(err error) (*NewerSmithError, bool) {
	var newer *NewerSmithError
	ok := errors.As(err, &newer)
	return newer, ok
}
