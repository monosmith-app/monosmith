package cipipeline

import (
	"strings"
	"sync"
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
		o.read(strings.TrimRight(o.pending[:i], "\r"))
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
