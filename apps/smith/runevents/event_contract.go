package runevents

import "time"

// Version is the version of the event shape. Additions are optional fields, so
// a client written for this version keeps working with a newer smith.
const Version = 1

// Kind names an event.
type Kind string

// The events of a run, in the order they happen.
const (
	RunStarted   Kind = "run.started"
	TaskQueued   Kind = "task.queued"
	TaskStarted  Kind = "task.started"
	TaskOutput   Kind = "task.output"
	TaskFinished Kind = "task.finished"
	RunFinished  Kind = "run.finished"
)

// Status is how a task ended.
type Status string

// The ways a task ends.
const (
	Passed  Status = "passed"
	Failed  Status = "failed"
	Cached  Status = "cached"
	Skipped Status = "skipped"
)

// Stream says where a line of output came from.
type Stream string

// The streams of task.output. A task's two streams arrive merged, as one
// terminal would show them, and are reported as Combined.
const (
	Stdout   Stream = "stdout"
	Stderr   Stream = "stderr"
	Combined Stream = "combined"
)

// Event is one fact about a run. Only the fields of its kind are set.
type Event struct {
	Event Kind `json:"event"`
	// Run identifies the run; Seq counts its events from 1, so a client can
	// detect a gap and order events that arrive from several runs.
	Run  string    `json:"run"`
	Seq  int       `json:"seq"`
	Time time.Time `json:"time"`

	// run.started
	Command string   `json:"command,omitempty"`
	Tasks   []string `json:"tasks,omitempty"`

	// task.queued, task.started, task.output, task.finished
	Task string `json:"task,omitempty"`

	// task.output
	Stream Stream `json:"stream,omitempty"`
	Line   string `json:"line,omitempty"`

	// task.finished
	Status     Status `json:"status,omitempty"`
	DurationMs int64  `json:"durationMs,omitempty"`
	CacheKey   string `json:"cacheKey,omitempty"`
	Reason     string `json:"reason,omitempty"`

	// run.finished
	OK      *bool    `json:"ok,omitempty"`
	Summary *Summary `json:"summary,omitempty"`
}

// Summary is the count of how the tasks of a run ended.
type Summary struct {
	Passed     int   `json:"passed"`
	Failed     int   `json:"failed"`
	Cached     int   `json:"cached"`
	Skipped    int   `json:"skipped"`
	DurationMs int64 `json:"durationMs"`
}

// Sink receives events. Emit may be called from several goroutines, but the
// events of one run arrive in order, one at a time.
type Sink interface {
	Emit(Event)
}

// SinkFunc makes a function a Sink.
type SinkFunc func(Event)

// Emit calls the function.
func (f SinkFunc) Emit(e Event) { f(e) }
