package runevents

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var runCounter atomic.Int64

// NewRunID returns an id no other run of this process has: the time smith
// started the run, and a count.
func NewRunID() string {
	return fmt.Sprintf("run-%s-%d", time.Now().UTC().Format("20060102T150405"), runCounter.Add(1))
}

// RunStream stamps the events of one run (its id, its sequence number, the
// time) and hands them to a sink one at a time, so they arrive in order. A nil
// RunStream drops everything, which is a run nobody listens to.
type RunStream struct {
	id   string
	sink Sink
	mu   sync.Mutex
	seq  int
}

// NewRunStream returns the stream of run id; with a nil sink it returns nil.
func NewRunStream(id string, sink Sink) *RunStream {
	if sink == nil {
		return nil
	}
	return &RunStream{id: id, sink: sink}
}

// Emit sets the run, the sequence and the time on e, and sends it.
func (r *RunStream) Emit(e Event) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	e.Run, e.Seq, e.Time = r.id, r.seq, time.Now().UTC()
	r.sink.Emit(e)
}

// Summarize counts how the results of a run ended.
func Summarize(passed, failed, cached, skipped int, took time.Duration) *Summary {
	return &Summary{Passed: passed, Failed: failed, Cached: cached, Skipped: skipped, DurationMs: took.Milliseconds()}
}
