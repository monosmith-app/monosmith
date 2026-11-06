package runevents

import (
	"encoding/json"
	"io"
	"sync"
)

// JSONLines writes each event as one line of JSON.
type JSONLines struct {
	mu  sync.Mutex
	out io.Writer
}

// NewJSONLines returns a sink that prints events to out as JSON lines.
func NewJSONLines(out io.Writer) *JSONLines { return &JSONLines{out: out} }

// Emit writes the event as a line. A broken output is not the run's problem.
func (j *JSONLines) Emit(e Event) {
	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	_, _ = j.out.Write(append(line, '\n'))
}
