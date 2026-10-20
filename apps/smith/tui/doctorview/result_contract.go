package doctorview

import "github.com/monosmith-app/monosmith/apps/smith/doctor"

// ProjectRef names a project and where it lives, so a finding that mentions
// either can be filed under it.
type ProjectRef struct {
	Name string
	Root string
}

// Result is what running doctor gave: the shell receives it as a message.
type Result struct {
	Report   doctor.Report
	Projects []ProjectRef
	Err      error
}

// Group is the findings that belong together: a project's, or the
// workspace's when Project is empty. Indexes point into the report's findings.
type Group struct {
	Project string
	Indexes []int
}
