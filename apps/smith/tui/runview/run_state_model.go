package runview

import (
	"sort"
	"strings"
	"time"
)

// Caps keep a long run from growing without bound.
const (
	maxRunLines  = 20000
	maxTaskLines = 5000
)

// Task is one task of a run.
type Task struct {
	ID      string
	Command string
	Status  Status
	// Started is when the task began; Duration is how long it took, once it has.
	Started  time.Time
	Duration time.Duration
	Reason   string
	// Log is what the task wrote.
	Log []string
	// order is the first-seen position, to keep the list stable.
	order int
}

// Run is the state of one `smith run` (or build, test, lint, typecheck).
type Run struct {
	// Command is the exact command line, without the program name.
	Command []string
	Tasks   []*Task
	// Lines is everything the process printed, in order.
	Lines   []string
	Started time.Time
	Ended   time.Time
	// Finished is true once the process has exited; Exit is its exit code.
	Finished  bool
	Exit      int
	Cancelled bool
	// total is the task count smith's own summary reported, or 0.
	total int
	index map[string]*Task
}

// NewRun starts the state of a run of command at now. planned lists the tasks
// expected to run, in order, when they are known up front; they start queued,
// which gives the progress bar its total.
func NewRun(command []string, planned []string, now time.Time) *Run {
	r := &Run{Command: command, Started: now, index: map[string]*Task{}}
	for _, id := range planned {
		r.task(id)
	}
	return r
}

func (r *Run) task(id string) *Task {
	if t, ok := r.index[id]; ok {
		return t
	}
	t := &Task{ID: id, order: len(r.Tasks)}
	r.index[id] = t
	r.Tasks = append(r.Tasks, t)
	return t
}

// Plan adds tasks expected to run that are not known yet, as queued, so the
// progress bar has its total. Tasks already seen keep their state.
func (r *Run) Plan(ids []string) {
	for _, id := range ids {
		r.task(id)
	}
}

// Task returns the task with an ID, or nil.
func (r *Run) Task(id string) *Task { return r.index[id] }

// Feed reads one line of output at time now and updates the run.
func (r *Run) Feed(line string, now time.Time) Event {
	ev := ParseLine(line)
	r.Lines = appendCapped(r.Lines, strings.TrimRight(line, "\r\n"), maxRunLines)
	switch ev.Kind {
	case Output:
		t := r.task(ev.Task)
		t.Log = appendCapped(t.Log, ev.Text, maxTaskLines)
	case Started:
		t := r.task(ev.Task)
		t.Status, t.Command, t.Started = Running, ev.Text, now
	case Finish:
		t := r.task(ev.Task)
		if t.Started.IsZero() {
			t.Started = now
		}
		t.Status, t.Duration, t.Reason = ev.Status, ev.Duration, ev.Text
		if ev.Status == Cached {
			t.Reason = ""
		}
		if ev.Status == Failed && t.Reason == "" {
			t.Reason = "failed"
		}
	case Stopped:
		t := r.task(ev.Task)
		t.Log = appendCapped(t.Log, "stopped", maxTaskLines)
	case Skip:
		t := r.task(ev.Task)
		t.Status, t.Reason = Skipped, ev.Text
	case Failure:
		t := r.task(ev.Task)
		t.Status = Failed
		if t.Reason == "" || t.Reason == "failed" {
			t.Reason = ev.Text
		}
	case Warned:
		t := r.task(ev.Task)
		t.Log = appendCapped(t.Log, "! "+ev.Text, maxTaskLines)
	case Summary:
		r.total = ev.Total
	}
	return ev
}

// End records that the process exited. A cancelled run marks what was still
// running as failed and what was still queued as skipped, so nothing is left
// spinning.
func (r *Run) End(exit int, cancelled bool, now time.Time) {
	r.Finished, r.Exit, r.Cancelled, r.Ended = true, exit, cancelled, now
	for _, t := range r.Tasks {
		switch {
		case t.Status == Running && (cancelled || exit != 0):
			t.Status, t.Reason, t.Duration = Failed, "interrupted", now.Sub(t.Started)
		case t.Status == Queued && cancelled:
			t.Status, t.Reason = Skipped, "cancelled"
		case t.Status == Queued && exit != 0:
			t.Status, t.Reason = Skipped, "the run ended before it started"
		}
	}
}

// Counts are how many tasks are in each state.
type Counts struct {
	Queued, Running, Cached, Passed, Failed, Skipped int
	// Total is every task known; Done is those with an outcome.
	Total, Done int
}

// Counts tallies the tasks.
func (r *Run) Counts() Counts {
	var c Counts
	for _, t := range r.Tasks {
		switch t.Status {
		case Queued:
			c.Queued++
		case Running:
			c.Running++
		case Cached:
			c.Cached++
		case Passed:
			c.Passed++
		case Failed:
			c.Failed++
		case Skipped:
			c.Skipped++
		}
	}
	c.Total = len(r.Tasks)
	if r.total > c.Total {
		c.Total = r.total
	}
	c.Done = c.Cached + c.Passed + c.Failed + c.Skipped
	return c
}

// OK reports whether the run finished with every task passed or cached.
func (r *Run) OK() bool {
	return r.Finished && r.Exit == 0 && !r.Cancelled && r.Counts().Failed == 0
}

// Elapsed is how long a task has been running (or ran) as of now.
func (r *Run) Elapsed(t *Task, now time.Time) time.Duration {
	switch {
	case t.Status.Finished():
		return t.Duration
	case t.Status == Running:
		return now.Sub(t.Started)
	}
	return 0
}

// RunTime is how long the run has taken, up to now while it is going.
func (r *Run) RunTime(now time.Time) time.Duration {
	if r.Finished {
		return r.Ended.Sub(r.Started)
	}
	return now.Sub(r.Started)
}

// rank orders the list: failures first, then what is going on, then the rest.
var rank = map[Status]int{Failed: 0, Running: 1, Queued: 2, Skipped: 3, Cached: 4, Passed: 5}

// Ordered is the tasks with failures first, then running, queued, skipped and
// finished ones; within a state, in the order they were first seen.
func (r *Run) Ordered() []*Task {
	out := append([]*Task(nil), r.Tasks...)
	sort.SliceStable(out, func(i, j int) bool {
		if rank[out[i].Status] != rank[out[j].Status] {
			return rank[out[i].Status] < rank[out[j].Status]
		}
		return out[i].order < out[j].order
	})
	return out
}

// CommandLine is the command as the user would type it.
func (r *Run) CommandLine() string { return "smith " + strings.Join(r.Command, " ") }

func appendCapped(lines []string, line string, limit int) []string {
	lines = append(lines, line)
	if len(lines) > limit {
		lines = append([]string(nil), lines[len(lines)-limit:]...)
	}
	return lines
}
