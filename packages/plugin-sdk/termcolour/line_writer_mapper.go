package termcolour

import (
	"bytes"
	"io"
	"sync"
)

// LineWriter passes tool output to out a line at a time through Line, so an
// escape sequence is never split and a line is painted whole. Call Flush when
// the tool ends, for a last line without a newline.
type LineWriter struct {
	out     io.Writer
	colour  bool
	mu      sync.Mutex
	pending []byte
}

// NewLineWriter returns a LineWriter.
//
// Parameters: out receives the lines; colour is the destination's decision.
// Returns: the writer.
func NewLineWriter(out io.Writer, colour bool) *LineWriter {
	return &LineWriter{out: out, colour: colour}
}

// Write implements io.Writer. It never fails: a write error of out is
// dropped, as a lost log line must not fail the tool.
//
// Parameters: p is a chunk of tool output.
// Returns: len(p) and nil.
func (w *LineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending = append(w.pending, p...)
	for {
		i := bytes.IndexByte(w.pending, '\n')
		if i < 0 {
			break
		}
		w.emit(w.pending[:i])
		w.pending = w.pending[i+1:]
	}
	return len(p), nil
}

// Flush writes a last line that had no newline.
func (w *LineWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.pending) > 0 {
		w.emit(w.pending)
		w.pending = nil
	}
}

func (w *LineWriter) emit(line []byte) {
	_, _ = io.WriteString(w.out, Line(string(bytes.TrimRight(line, "\r")), w.colour)+"\n")
}
