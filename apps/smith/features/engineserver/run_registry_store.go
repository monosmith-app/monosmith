package engineserver

import (
	"context"
	"fmt"
	"sync"
)

// runRegistry knows the runs in flight, so a cancel can reach them.
type runRegistry struct {
	mu   sync.Mutex
	runs map[string]context.CancelFunc
}

// begin registers run id, or fails when it is already in flight.
func (r *runRegistry) begin(id string, cancel context.CancelFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.runs == nil {
		r.runs = map[string]context.CancelFunc{}
	}
	if _, busy := r.runs[id]; busy {
		return fmt.Errorf("run %q is already in flight", id)
	}
	r.runs[id] = cancel
	return nil
}

func (r *runRegistry) end(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, id)
}

// cancel stops run id; false when no such run is in flight.
func (r *runRegistry) cancel(id string) bool {
	r.mu.Lock()
	cancel, ok := r.runs[id]
	r.mu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

func (r *runRegistry) cancelAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cancel := range r.runs {
		cancel()
	}
}
