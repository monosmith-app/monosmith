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

// Extra is a check doctor does not run itself (the vertical-slice layout),
// listed under the findings as a row that runs it now. Presents is the id of
// the catalogue action whose guided form runs it ("slices check").
type Extra struct {
	Title, Hint, Presents string
}

// Group is the findings that belong together: a project's, or the
// workspace's when Project is empty. Indexes point into the report's findings.
// Counted, when set, points at every finding of the group before the filters
// hid any: the heading counts those, not only the rows drawn.
type Group struct {
	Project string
	Indexes []int
	Counted []int
}
