package conformance

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// cancelBound is how long a plugin that advertised cancel has to answer a
// request the host cancelled.
var cancelBound = 10 * time.Second

// unknownRequest is an id no request ever had.
const unknownRequest = 1 << 40

// probeOf finds the cheapest call the plugin answers without doing work: a
// method that takes a list of projects or manifests, asked about none.
func probeOf(c *host.Client) (func(ctx context.Context) error, bool) {
	probes := []struct {
		method string
		params any
		result any
	}{
		{protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{}}, &protocol.DiscoverResult{}},
		{protocol.MethodContribute, protocol.ContributeParams{Projects: []protocol.Project{}}, &protocol.ContributeResult{}},
		{protocol.MethodCI, protocol.CIParams{Projects: []protocol.Project{}}, &protocol.CIResult{}},
		{protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{}}, &protocol.DoctorResult{}},
	}
	for _, p := range probes {
		if c.Supports(p.method) {
			return func(ctx context.Context) error {
				return c.CallContext(ctx, p.method, p.params, p.result, 2*time.Minute)
			}, true
		}
	}
	return nil, false
}

// checkScheduling checks what a plugin says about how it is called: the
// concurrency it declares must be one the protocol defines and be accepted,
// and the cancel it advertises must stop a request within a bound and leave
// the plugin serving.
func checkScheduling(c *host.Client, r *Report) {
	info := c.Info
	r.add("declares a concurrency the protocol defines", protocol.ConcurrencyValid(info.Concurrency),
		"concurrency %d is below -1; it must be absent, 1 or more, or -1 for unlimited", info.Concurrency)
	probe, ok := probeOf(c)
	if !ok {
		return
	}
	checkConcurrency(c, probe, r)
	if info.Capabilities.Cancel {
		checkCancel(c, probe, r)
	}
}

// checkConcurrency sends requests together. A plugin that declares nothing is
// sent them one at a time and must answer them all; one that declares a
// concurrency is sent that many at once (at most four, to keep the check
// cheap) and must answer them all.
func checkConcurrency(c *host.Client, probe func(context.Context) error, r *Report) {
	limit := host.InFlightLimit(c.Info)
	together := min(limit, 4)
	name := fmt.Sprintf("accepts %d requests at the same time", together)
	if limit == 1 {
		together = 3
		name = "answers 3 requests sent together, one at a time"
	}
	errs := make([]error, together)
	var wg sync.WaitGroup
	for i := range together {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = probe(context.Background())
		}()
	}
	wg.Wait()
	err := errors.Join(errs...)
	peak := c.PeakInFlight()
	r.add(name, err == nil && peak <= limit, "%v (the host had %d in flight, the plugin declared %d)", err, peak, limit)
}

// checkCancel cancels a request, then a request that does not exist.
func checkCancel(c *host.Client, probe func(context.Context) error, r *Report) {
	ctx, cancel := context.WithCancel(context.Background())
	failed := make(chan error, 1)
	began := time.Now()
	go func() { failed <- probe(ctx) }()
	time.Sleep(5 * time.Millisecond) // let the request leave; a cancel that wins the race is just as good an answer
	cancel()
	var err error
	select {
	case err = <-failed:
	case <-time.After(cancelBound + c.CancelGrace()):
		err = fmt.Errorf("no answer within %s of the cancel", cancelBound)
	}
	stopped := err == nil || (errors.Is(err, context.Canceled) && !host.WasKilled(err))
	r.add("answers a cancelled request with its result or the cancelled error", stopped && time.Since(began) < cancelBound,
		"%v (a plugin that advertised cancel answers the request, with %d if it stopped, within %s)", err, protocol.CodeRequestCancelled, cancelBound)

	notifyErr := c.Notify(protocol.MethodCancel, protocol.CancelParams{ID: unknownRequest})
	followUp := probe(context.Background())
	r.add("ignores a cancel for a request it does not know", notifyErr == nil && followUp == nil, "%v, then %v", notifyErr, followUp)
}
