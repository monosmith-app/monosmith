package serve

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// unlimitedWorkers bounds the pool of a plugin that declared unlimited
// concurrency: the host decides how many requests it sends, but the plugin
// still never runs more goroutines than this.
const unlimitedWorkers = 1024

// maxQueued bounds the requests waiting for a worker. The host never has more
// in flight than the plugin declared, so a queue this long means a host that
// breaks the protocol; its further requests are refused.
const maxQueued = 1024

// pooledRequest is one request from the moment it is read until it is
// answered.
type pooledRequest struct {
	req       protocol.Request
	ctx       context.Context
	stop      context.CancelFunc
	cancelled atomic.Bool
}

// requestPool is the state of the serve loop: the requests in flight, the
// ones waiting for a worker, and the bounded set of workers running them.
// Requests are started in the order they arrive; they may finish in any order.
type requestPool struct {
	mu       sync.Mutex
	limit    int
	running  int
	queue    []*pooledRequest
	inflight map[int64]*pooledRequest
	workers  sync.WaitGroup
}

func newRequestPool() *requestPool {
	return &requestPool{limit: 1, inflight: map[int64]*pooledRequest{}}
}

// workersFor is the pool size a declared concurrency asks for.
func workersFor(declared int) int {
	limit, unlimited := protocol.ConcurrencyLimit(declared)
	if unlimited {
		return unlimitedWorkers
	}
	return limit
}

// setLimit changes how many workers may run at once. Raising it starts
// waiting requests; lowering it lets running ones finish.
func (p *requestPool) setLimit(n int, run func(*pooledRequest)) {
	p.mu.Lock()
	p.limit = max(n, 1)
	var started []*pooledRequest
	for p.running < p.limit && len(p.queue) > 0 {
		started = append(started, p.queue[0])
		p.queue = p.queue[1:]
		p.running++
		p.workers.Add(1)
	}
	p.mu.Unlock()
	for _, pr := range started {
		go p.work(pr, run)
	}
}

// admission is what submit decided.
type admission int

const (
	admitted admission = iota
	refusedDuplicate
	refusedFull
)

// submit takes a request: it runs at once on a free worker, or waits its turn.
func (p *requestPool) submit(req protocol.Request, run func(*pooledRequest)) admission {
	p.mu.Lock()
	if _, busy := p.inflight[*req.ID]; busy {
		p.mu.Unlock()
		return refusedDuplicate
	}
	if p.running >= p.limit && len(p.queue) >= maxQueued {
		p.mu.Unlock()
		return refusedFull
	}
	ctx, stop := context.WithCancel(context.Background())
	pr := &pooledRequest{req: req, ctx: ctx, stop: stop}
	p.inflight[*req.ID] = pr
	if p.running >= p.limit {
		p.queue = append(p.queue, pr)
		p.mu.Unlock()
		return admitted
	}
	p.running++
	p.workers.Add(1)
	p.mu.Unlock()
	go p.work(pr, run)
	return admitted
}

// work runs first, then whatever waits, and gives its slot back when none does.
func (p *requestPool) work(first *pooledRequest, run func(*pooledRequest)) {
	defer p.workers.Done()
	for pr := first; pr != nil; pr = p.next() {
		run(pr)
	}
}

func (p *requestPool) next() *pooledRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.queue) > 0 && p.running <= p.limit {
		pr := p.queue[0]
		p.queue = p.queue[1:]
		return pr
	}
	p.running--
	return nil
}

// release forgets an answered request; call it before the answer is written,
// so a host that reuses the id right away is not refused as a duplicate.
func (p *requestPool) release(pr *pooledRequest) {
	p.mu.Lock()
	delete(p.inflight, *pr.req.ID)
	p.mu.Unlock()
	pr.stop()
}

// cancel stops the request with an id. A request that had not started is taken
// out of the queue and returned, because no worker will ever answer it. An
// unknown id is ignored.
func (p *requestPool) cancel(id int64) (waiting *pooledRequest) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pr, ok := p.inflight[id]
	if !ok {
		return nil
	}
	pr.cancelled.Store(true)
	pr.stop()
	for i, queued := range p.queue {
		if queued == pr {
			p.queue = append(p.queue[:i], p.queue[i+1:]...)
			delete(p.inflight, id)
			return pr
		}
	}
	return nil
}

// cancelAll stops every request. The ones that had not started are returned
// to be answered.
func (p *requestPool) cancelAll() (waiting []*pooledRequest) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, pr := range p.inflight {
		pr.cancelled.Store(true)
		pr.stop()
	}
	waiting = p.queue
	for _, pr := range waiting {
		delete(p.inflight, *pr.req.ID)
	}
	p.queue = nil
	return waiting
}

// wait blocks until no worker is running.
func (p *requestPool) wait() { p.workers.Wait() }
