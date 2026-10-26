package adoptwizard

import (
	"bytes"
	"io"
	"strings"
	"time"
)

// checkedStep is one step that finished, and how long it took.
type checkedStep struct {
	Name string
	Took time.Duration
}

// stepTimer sits between a job and the shell: it passes everything on and
// notes, from the "… name" and "✔ name" lines the steps write, how long each
// step took.
type stepTimer struct {
	out     io.Writer
	now     func() time.Time
	pending []byte
	started map[string]time.Time
	done    []checkedStep
}

func newStepTimer(out io.Writer) *stepTimer {
	return &stepTimer{out: out, now: time.Now, started: map[string]time.Time{}}
}

// Write passes p on and reads the lines in it.
func (t *stepTimer) Write(p []byte) (int, error) {
	t.pending = append(t.pending, p...)
	for {
		i := bytes.IndexByte(t.pending, '\n')
		if i < 0 {
			break
		}
		t.see(strings.TrimRight(string(t.pending[:i]), "\r"))
		t.pending = t.pending[i+1:]
	}
	return t.out.Write(p)
}

func (t *stepTimer) see(line string) {
	if name, ok := strings.CutPrefix(line, "… "); ok {
		t.started[name] = t.now()
	} else if name, ok := strings.CutPrefix(line, "✔ "); ok {
		if at, seen := t.started[name]; seen {
			t.done = append(t.done, checkedStep{Name: name, Took: t.now().Sub(at)})
		}
	}
}

// Steps are the steps that finished, in order.
func (t *stepTimer) Steps() []checkedStep { return t.done }
