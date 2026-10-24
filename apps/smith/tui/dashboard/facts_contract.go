package dashboard

// Facts is what the page is drawn from; the shell fills it from what it has
// already loaded, so drawing never waits for anything.
type Facts struct {
	// Checked is true once doctor has run; Failed, Warnings and Passed count its findings.
	Checked                  bool
	CheckErr                 string
	Failed, Warnings, Passed int
	// Fixable is how many fixes doctor can make; the fix wizard opens on them.
	Fixable int
	// ProjectsLoaded is true once the project list arrived.
	ProjectsLoaded     bool
	Projects, Affected int
	// Last is the last command that ran, nil before the first.
	Last *Run
	// DepsChecked is true once the Deps view looked for newer releases.
	DepsChecked bool
	Outdated    int
}

// Run is how the last command ended.
type Run struct {
	// Command is the command line without the program name.
	Command   string
	OK        bool
	Cancelled bool
	// Took is the duration, already formatted.
	Took string
}

// Next is the one suggested action: a sentence naming its key, and the id of
// the command the key runs ("" when there is nothing to press yet).
type Next struct {
	Text      string
	CommandID string
}
