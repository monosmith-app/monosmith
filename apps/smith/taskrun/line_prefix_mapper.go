package taskrun

import (
	"bytes"
	"io"
	"sync"
)

// prefixWriter writes complete lines to out, each starting with prefix, and
// keeps an unprefixed copy of everything for the cache log. Lines from tasks
// running in parallel never interleave mid-line.
type prefixWriter struct {
	prefix  string
	out     io.Writer
	outMu   *sync.Mutex
	pending []byte
	log     bytes.Buffer
	// hold keeps the prefixed lines in held instead of writing them; Release
	// writes them. A quiet run holds a task's output until it knows the task failed.
	hold bool
	held bytes.Buffer
	// onLine, when set, receives every line without its prefix, for the event stream.
	onLine func(string)
}

func newPrefixWriter(prefix string, out io.Writer, mu *sync.Mutex) *prefixWriter {
	return &prefixWriter{prefix: prefix, out: out, outMu: mu}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.log.Write(p)
	w.pending = append(w.pending, p...)
	for {
		i := bytes.IndexByte(w.pending, '\n')
		if i < 0 {
			break
		}
		w.emit(w.pending[:i+1])
		w.pending = w.pending[i+1:]
	}
	return len(p), nil
}

// Flush writes a trailing partial line.
func (w *prefixWriter) Flush() {
	if len(w.pending) > 0 {
		w.emit(append(w.pending, '\n'))
		w.pending = nil
	}
}

func (w *prefixWriter) emit(line []byte) {
	w.outMu.Lock()
	defer w.outMu.Unlock()
	var dest io.Writer = w.out
	if w.hold {
		dest = &w.held
	}
	_, _ = io.WriteString(dest, w.prefix)
	_, _ = dest.Write(bytes.TrimRight(line, "\r\n"))
	_, _ = io.WriteString(dest, "\n")
	if w.onLine != nil {
		w.onLine(string(bytes.TrimRight(line, "\r\n")))
	}
}

// Release writes the lines a quiet run held back, for a task that failed.
func (w *prefixWriter) Release() {
	w.outMu.Lock()
	defer w.outMu.Unlock()
	_, _ = w.out.Write(w.held.Bytes())
	w.held.Reset()
}

// Replay writes a cached log through the same prefixing.
func Replay(prefix string, log []byte, out io.Writer, mu *sync.Mutex, onLine func(string)) {
	w := newPrefixWriter(prefix, out, mu)
	w.onLine = onLine
	_, _ = w.Write(log)
	w.Flush()
}
