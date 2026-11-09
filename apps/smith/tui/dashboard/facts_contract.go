package dashboard

import "time"

// Facts is what the page is drawn from; the shell fills it from what it has
// already loaded, so drawing never waits for anything.
type Facts struct {
	// Checked is true once doctor has run; Failed, Warnings and Passed count its findings.
	Checked bool
	// Waited is how long the first doctor run has taken so far.
	Waited                   time.Duration
	CheckErr                 string
	Failed, Warnings, Passed int
	// Fixable is how many fixes doctor can make; the fix wizard opens on them.
	Fixable int
	// Problems name the failing checks, one line each, the first one first.
	Problems []string
	// WarningNotes name the checks that warn, one line each, the first one first.
	WarningNotes []string
	// ProjectsLoaded is true once the project list arrived.
	ProjectsLoaded     bool
	Projects, Affected int
	// Languages is the language of each project, to say what the workspace holds.
	Languages []string
	// Last is the last command that ran, nil before the first.
	Last *Run
	// Recent are the runs of the session that ended, the newest first.
	Recent []Run
	// DepsChecked is true once the Deps view looked for newer releases.
	DepsChecked bool
	Outdated    int
	// Release is what the next release would publish, nil until the plan is
	// read and while there is nothing to release.
	Release *Ready
	// SettingsTip is true until the person has opened the Settings tab: Home points at it.
	SettingsTip bool
}

// Run is how a command ended.
type Run struct {
	// Command is the command line without the program name.
	Command string
	// Task is the id of the first task that failed ("orders:build"), "" when
	// the run did not fail in a task.
	Task      string
	OK        bool
	Cancelled bool
	// Took is the duration, already formatted.
	Took string
	// Repeatable is true when running the command again is allowed.
	Repeatable bool
	// Index is the place of the run among the recent ones, 0 for the newest; a
	// click on it opens that run.
	Index int
}

// Ready is the next release in a few facts: how many packages leave, and the
// one that is named in the sentence.
type Ready struct {
	// Packages is how many packages the release publishes.
	Packages int
	// Name, From and To are the package named: its version now ("" before its
	// first release) and the one it gets.
	Name, From, To string
	// Commits is how many commits the named package releases (0 when unknown).
	Commits int
}
