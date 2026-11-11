package taskrun

import (
	"bytes"
	"io"
	"sync"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// prefixWriter writes complete lines to out, each starting with prefix, and
// keeps an unprefixed copy of everything for the cache log (Plain). Lines from
// tasks running in parallel never interleave mid-line.
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
	// onLine, when set, receives every line without its prefix and without
	// escape codes, for the event stream.
	onLine func(string)
	// colour is the destination's decision: on, a line keeps only its SGR
	// codes and smith paint applies; off, every escape code goes.
	colour bool
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
	dest := io.Writer(w.out)
	if w.hold {
		dest = &w.held
	}
	text := string(bytes.TrimRight(line, "\r\n"))
	_, _ = io.WriteString(dest, w.prefix)
	_, _ = io.WriteString(dest, termcolour.Line(text, w.colour))
	_, _ = io.WriteString(dest, "\n")
	if w.onLine != nil {
		w.onLine(termcolour.Strip(text))
	}
}

// Plain is the cache log: everything the task wrote, without escape codes.
// The cache stores the log plain, so an entry from any machine replays as
// plain text, and smith paint colours it again for a destination with colour.
func (w *prefixWriter) Plain() []byte { return []byte(termcolour.Strip(w.log.String())) }

// Release writes the lines a quiet run held back, for a task that failed.
func (w *prefixWriter) Release() {
	w.outMu.Lock()
	defer w.outMu.Unlock()
	_, _ = w.out.Write(w.held.Bytes())
	w.held.Reset()
}

// Replay writes a cached log through the same prefixing. A log that holds
// escape codes (stored by an older smith) is filtered like live output.
//
// Parameters: prefix starts every line; log is the cached log; out and mu are
// the destination and its lock; onLine, when not nil, receives each plain
// line; colour is the destination's colour decision.
func Replay(prefix string, log []byte, out io.Writer, mu *sync.Mutex, onLine func(string), colour bool) {
	w := newPrefixWriter(prefix, out, mu)
	w.onLine = onLine
	w.colour = colour
	_, _ = w.Write(log)
	w.Flush()
}
