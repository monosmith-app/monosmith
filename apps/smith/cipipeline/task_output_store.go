package cipipeline

import (
	"github.com/monosmith-app/monosmith/apps/smith/verifyfacts"
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// taskTailLines is how many of the last lines of each task are kept, to show
// the cause of a failure outside the folded output.
const taskTailLines = 15

// taskOutputs listens to the prefixed lines `smith run` prints
// ("<project>:<task> │ <line>") and keeps, per task, the end of its output and
// what its test runner said. Tasks run in parallel, so it is safe for
// concurrent writes.
type taskOutputs struct {
	mu      sync.Mutex
	pending string
	tasks   map[string]*taskOutput
}

type taskOutput struct {
	tail  []string
	tests testCounter
	// facts and lint keep what the runner and the linter said, for the details of verify.
	facts verifyfacts.TestCounter
	lint  verifyfacts.LintCounter
}

func newTaskOutputs() *taskOutputs { return &taskOutputs{tasks: map[string]*taskOutput{}} }

// Write reads whole lines; a trailing partial line waits for its end.
func (o *taskOutputs) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.pending += string(p)
	for {
		i := strings.IndexByte(o.pending, '\n')
		if i < 0 {
			break
		}
		// The tail, the test counts and the job summary are read and
		// printed outside the group: they stay plain whatever the log shows.
		o.read(termcolour.Strip(strings.TrimRight(o.pending[:i], "\r")))
		o.pending = o.pending[i+1:]
	}
	return len(p), nil
}

func (o *taskOutputs) read(line string) {
	id, text, ok := strings.Cut(line, " │ ")
	if !ok || strings.ContainsAny(id, " \t") {
		return
	}
	t := o.tasks[id]
	if t == nil {
		t = &taskOutput{}
		o.tasks[id] = t
	}
	t.tail = append(t.tail, text)
	if len(t.tail) > taskTailLines {
		t.tail = t.tail[1:]
	}
	t.tests.add(text)
	t.facts.Add(text)
	t.lint.Add(text)
}

// Tail is the last lines of a task's output, or "".
func (o *taskOutputs) Tail(id string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	if t := o.tasks[id]; t != nil {
		return strings.Join(t.tail, "\n")
	}
	return ""
}

// Tests is what the task's test runner said ("42 passed"), or "".
func (o *taskOutputs) Tests(id string) (summary string, passed int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if t := o.tasks[id]; t != nil {
		return t.tests.summary(), t.tests.passed
	}
	return "", 0
}

// TestFacts is what the task's test runner counted: passed and failed, what they
// count ("tests" or "packages"), and whether the runner said anything readable.
func (o *taskOutputs) TestFacts(id string) (passed, failed int, unit string, known bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if t := o.tasks[id]; t != nil {
		return t.facts.Result()
	}
	return 0, 0, "", false
}

// LintIssues is the number of issues the task's linter reported, and whether it said.
func (o *taskOutputs) LintIssues(id string) (issues int, known bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if t := o.tasks[id]; t != nil {
		return t.lint.Issues()
	}
	return 0, false
}
