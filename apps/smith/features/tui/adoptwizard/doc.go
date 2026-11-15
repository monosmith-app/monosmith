// Package adoptwizard is the dialog that brings a repository under smith: it
// shows what was found (projects, the old tooling), asks how far to go
// (conservative or full), for the full level asks the four choices, shows the
// whole plan, converts, asks about each group of removals (yes, no, skip),
// and runs the closing install and checks, step by step.
//
// It is a state machine over the fix wizard: every screen is a
// fixwizard.Wizard (questions, a scrollable review, progress, a failure with
// its output and Retry), so keys, mouse and drawing are the same as in the
// Doctor fix dialog. It touches nothing itself. A step that takes time comes
// out as a Job the shell runs off the UI goroutine; the shell reports its
// progress and its end back (Progress, Done). What a job calls is the Flow,
// which adoptflow.Session implements, so tests drive the whole walk with a
// fake.
package adoptwizard
