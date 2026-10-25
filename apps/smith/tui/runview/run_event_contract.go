package runview

import "time"

// Status is where a task is in the run.
type Status int

// The states of a task. Order matters: it is the order of the summary counts.
const (
	Queued Status = iota
	Running
	Cached
	Passed
	Failed
	Skipped
)

// String is the word shown beside the symbol, so colour is never the only signal.
func (s Status) String() string {
	return [...]string{"queued", "running", "cached", "passed", "failed", "skipped"}[s]
}

// Finished reports whether the task has an outcome.
func (s Status) Finished() bool { return s >= Cached }

// EventKind is what a line of smith's output says.
type EventKind int

// The kinds of line.
const (
	// Plain is a line that says nothing about a task: "affected: ...", "nothing to run".
	Plain EventKind = iota
	// Output is a line a task wrote.
	Output
	// Started is "<task> ▶ <command>".
	Started
	// Finish is "<task> ✔ <duration>", "<task> ✔ from cache" or "<task> ✖ <reason>".
	Finish
	// Stopped is "<task> ■ stopped" (a served task shut down).
	Stopped
	// Warned is "<task> ! <message>".
	Warned
	// Skip is "  ↷ <task> <reason>" in the closing summary.
	Skip
	// Failure is "  ✖ <task> <reason>" in the closing summary.
	Failure
	// Summary is "✔ 3/4 tasks succeeded (...) in 1.2s".
	Summary
	// Plan is the list of tasks a run will start (the event stream's
	// run.started); Queue is one of them (task.queued).
	Plan
	Queue
)

// Event is one parsed line.
type Event struct {
	Kind EventKind
	// Task is the task's ID ("web:build"), empty for Plain.
	Task string
	// Text is the output line, the command, the reason or the message.
	Text string
	// Status is the outcome carried by a Finish, Skip or Failure.
	Status Status
	// Duration is how long a task took (Finish).
	Duration time.Duration
	// Done and Total are the counts in a Summary line; OK says whether it
	// is the success form.
	Done, Total int
	OK          bool
	// Tasks is the plan of a run (Plan).
	Tasks []string
	// Display is the line as the log shows it, when the line read was not the
	// text smith prints (an event of the stream): empty means the line itself.
	Display string
	// Hidden is set for an event that says something to the state but has no
	// line of its own in the log.
	Hidden bool
}
