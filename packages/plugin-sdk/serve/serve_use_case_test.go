package serve_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/transport"
)

const bound = 5 * time.Second

// state is what a lab plugin shares across the copies Serve makes of it.
type state struct {
	concurrency int
	started     chan string
	release     chan struct{}
	mu          sync.Mutex
	active      int
	peak        int
	cancelled   int
}

// lab is a plugin whose discover does what the first manifest name says:
// "block-*" waits to be released (or cancelled), "slow-*" sleeps a moment,
// anything else answers at once. The answer's project is the manifest name.
type lab struct {
	*state
	ctx context.Context
}

func newLab(concurrency int) lab {
	return lab{state: &state{concurrency: concurrency, started: make(chan string, 4096), release: make(chan struct{})}}
}

func (l lab) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "lab", Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"lab.json"}, Concurrency: l.concurrency}, nil
}

func (l lab) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	name := p.Manifests[0]
	l.mu.Lock()
	l.active++
	l.peak = max(l.peak, l.active)
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		l.active--
		l.mu.Unlock()
	}()
	l.started <- name
	switch {
	case strings.HasPrefix(name, "block"):
		var done <-chan struct{}
		if l.ctx != nil {
			done = l.ctx.Done()
		}
		select {
		case <-l.release:
		case <-done:
			l.mu.Lock()
			l.cancelled++
			l.mu.Unlock()
			return protocol.DiscoverResult{}, l.ctx.Err()
		}
	case strings.HasPrefix(name, "slow"):
		time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)
	case strings.HasPrefix(name, "fail"):
		return protocol.DiscoverResult{}, errors.New("the manifest is broken")
	case strings.HasPrefix(name, "stubborn"):
		// Finishes its work even when cancelled, and says so with a result.
		<-l.release
	}
	return protocol.DiscoverResult{Projects: []protocol.Project{{Name: name, Root: name, Language: "lab", Kind: protocol.ProjectLib}}}, nil
}

// cancellableLab is a lab that advertises cancel.
type cancellableLab struct{ lab }

func (c cancellableLab) WithContext(ctx context.Context) serve.Plugin {
	c.ctx = ctx
	return c
}

// wire is the host's end of a Serve running in this process.
type wire struct {
	t         *testing.T
	stream    *transport.Stream
	closeIn   func()
	responses chan protocol.Response
	done      chan struct{}
	serveErr  error
}

func start(t *testing.T, p serve.Plugin) *wire {
	t.Helper()
	pluginIn, hostOut := io.Pipe()
	hostIn, pluginOut := io.Pipe()
	w := &wire{t: t, stream: transport.NewStream(hostIn, hostOut), closeIn: func() { _ = hostOut.Close() }, responses: make(chan protocol.Response, 8192), done: make(chan struct{})}
	go func() {
		w.serveErr = serve.Serve(p, pluginIn, pluginOut)
		_ = pluginOut.Close()
		close(w.done)
	}()
	go func() {
		for {
			body, err := w.stream.Read()
			if err != nil {
				close(w.responses)
				return
			}
			var resp protocol.Response
			if err := json.Unmarshal(body, &resp); err != nil {
				t.Errorf("malformed response %s", body)
				continue
			}
			w.responses <- resp
		}
	}()
	t.Cleanup(func() {
		w.closeIn()
		select {
		case <-w.done:
		case <-time.After(bound):
			t.Error("Serve did not return after the host closed the stream")
		}
	})
	return w
}

func (w *wire) request(id int64, method string, params any) {
	w.t.Helper()
	raw, _ := json.Marshal(params)
	body, _ := json.Marshal(protocol.Request{JSONRPC: "2.0", ID: &id, Method: method, Params: raw})
	if err := w.stream.Write(body); err != nil {
		w.t.Fatal(err)
	}
}

func (w *wire) notify(method string, params any) {
	w.t.Helper()
	raw, _ := json.Marshal(params)
	body, _ := json.Marshal(protocol.Request{JSONRPC: "2.0", Method: method, Params: raw})
	if err := w.stream.Write(body); err != nil {
		w.t.Fatal(err)
	}
}

func (w *wire) discover(id int64, name string) {
	w.t.Helper()
	w.request(id, protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{name}})
}

func (w *wire) initialize() protocol.InitializeResult {
	w.t.Helper()
	w.request(1000000, protocol.MethodInitialize, protocol.InitializeParams{ProtocolVersions: []int{protocol.Version}})
	var res protocol.InitializeResult
	resp := w.next()
	if resp.Error != nil {
		w.t.Fatalf("initialize: %v", resp.Error)
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		w.t.Fatal(err)
	}
	return res
}

func (w *wire) next() protocol.Response {
	w.t.Helper()
	select {
	case resp, ok := <-w.responses:
		if !ok {
			w.t.Fatal("the plugin closed the stream")
		}
		return resp
	case <-time.After(bound):
		w.t.Fatal("no response within the bound")
		return protocol.Response{}
	}
}

func (w *wire) silent(d time.Duration) {
	w.t.Helper()
	select {
	case resp := <-w.responses:
		w.t.Fatalf("unexpected response %+v", resp)
	case <-time.After(d):
	}
}

func (l lab) waitStarted(t *testing.T, n int) []string {
	t.Helper()
	var got []string
	for len(got) < n {
		select {
		case name := <-l.started:
			got = append(got, name)
		case <-time.After(bound):
			t.Fatalf("only %v of %d handlers started", got, n)
		}
	}
	return got
}

func (l lab) peakActive() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.peak
}

func project(t *testing.T, resp protocol.Response) string {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("response %d: %v", *resp.ID, resp.Error)
	}
	var res protocol.DiscoverResult
	if err := json.Unmarshal(resp.Result, &res); err != nil || len(res.Projects) != 1 {
		t.Fatalf("response %d: %s (%v)", *resp.ID, resp.Result, err)
	}
	return res.Projects[0].Name
}

func TestAPluginThatDeclaresNothingHandlesOneRequestAtATimeInOrder(t *testing.T) {
	p := newLab(0)
	w := start(t, p)
	res := w.initialize()
	if res.Concurrency != 0 || res.Capabilities.Cancel {
		t.Fatalf("a plain plugin declares nothing: %+v", res)
	}
	for id := int64(1); id <= 6; id++ {
		w.discover(id, fmt.Sprintf("slow-%d", id))
	}
	for id := int64(1); id <= 6; id++ {
		resp := w.next()
		if *resp.ID != id || project(t, resp) != fmt.Sprintf("slow-%d", id) {
			t.Fatalf("response %d arrived where %d was due", *resp.ID, id)
		}
	}
	if peak := p.peakActive(); peak != 1 {
		t.Fatalf("a default plugin ran %d handlers at once", peak)
	}
}

func TestADeclaredConcurrencyRunsThatManyHandlersAtOnceAndNoMore(t *testing.T) {
	p := newLab(3)
	w := start(t, p)
	if res := w.initialize(); res.Concurrency != 3 {
		t.Fatalf("concurrency %d", res.Concurrency)
	}
	for id := int64(1); id <= 8; id++ {
		w.discover(id, fmt.Sprintf("block-%d", id))
	}
	p.waitStarted(t, 3) // the pool is full: the other five wait their turn
	time.Sleep(50 * time.Millisecond)
	if peak := p.peakActive(); peak != 3 {
		t.Fatalf("%d handlers ran at once, want exactly the declared 3", peak)
	}
	close(p.release)
	seen := map[int64]bool{}
	for range 8 {
		resp := w.next()
		project(t, resp)
		seen[*resp.ID] = true
	}
	if len(seen) != 8 || p.peakActive() != 3 {
		t.Fatalf("answered %v with a peak of %d", seen, p.peakActive())
	}
}

func TestResponsesMayComeOutOfOrder(t *testing.T) {
	p := newLab(2)
	w := start(t, p)
	w.initialize()
	w.discover(1, "block-first")
	w.discover(2, "quick")
	if resp := w.next(); *resp.ID != 2 || project(t, resp) != "quick" {
		t.Fatalf("the quick request must not wait for the blocked one: %+v", resp)
	}
	close(p.release)
	if resp := w.next(); *resp.ID != 1 {
		t.Fatalf("the blocked request is answered last: %+v", resp)
	}
}

func TestCancelStopsARunningRequestWithTheCancelledError(t *testing.T) {
	p := cancellableLab{newLab(0)}
	w := start(t, p)
	if res := w.initialize(); !res.Capabilities.Cancel {
		t.Fatal("a Cancellable plugin must advertise the cancel capability")
	}
	w.discover(1, "block-a")
	p.waitStarted(t, 1)
	began := time.Now()
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 1})
	resp := w.next()
	if resp.Error == nil || resp.Error.Code != protocol.CodeRequestCancelled || *resp.ID != 1 {
		t.Fatalf("want request cancelled for 1, got %+v", resp)
	}
	if time.Since(began) > time.Second {
		t.Fatalf("the cancel took %s to take effect", time.Since(began))
	}
	w.discover(2, "quick")
	if project(t, w.next()) != "quick" {
		t.Fatal("the plugin must go on serving after a cancel")
	}
}

func TestCancelOfAnUnknownOrFinishedRequestIsIgnored(t *testing.T) {
	p := cancellableLab{newLab(0)}
	w := start(t, p)
	w.initialize()
	w.discover(1, "quick")
	project(t, w.next())
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 1})
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 424242})
	w.notify(protocol.MethodCancel, json.RawMessage(`"not params"`))
	w.silent(50 * time.Millisecond)
	w.discover(2, "quick")
	project(t, w.next())
}

func TestARequestFinishingDespiteTheCancelAnswersWithItsResult(t *testing.T) {
	p := cancellableLab{newLab(0)}
	w := start(t, p)
	w.initialize()
	w.discover(1, "stubborn")
	p.waitStarted(t, 1)
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 1})
	w.silent(50 * time.Millisecond)
	close(p.release)
	if project(t, w.next()) != "stubborn" {
		t.Fatal("work that completes is answered with its result")
	}
}

func TestACancelledRequestThatHadNotStartedIsAnsweredAtOnce(t *testing.T) {
	p := cancellableLab{newLab(1)}
	w := start(t, p)
	w.initialize()
	w.discover(1, "block-running")
	w.discover(2, "block-waiting")
	p.waitStarted(t, 1)
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 2})
	resp := w.next()
	if *resp.ID != 2 || resp.Error == nil || resp.Error.Code != protocol.CodeRequestCancelled {
		t.Fatalf("the waiting request is refused, got %+v", resp)
	}
	close(p.release)
	if resp := w.next(); *resp.ID != 1 || project(t, resp) != "block-running" {
		t.Fatalf("the running request is untouched: %+v", resp)
	}
	select {
	case name := <-p.started:
		t.Fatalf("the cancelled request still started: %s", name)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestAPluginThatCannotBeCancelledIgnoresCancel(t *testing.T) {
	p := newLab(0)
	w := start(t, p)
	w.initialize()
	w.discover(1, "stubborn")
	p.waitStarted(t, 1)
	w.notify(protocol.MethodCancel, protocol.CancelParams{ID: 1})
	w.silent(50 * time.Millisecond)
	close(p.release)
	if project(t, w.next()) != "stubborn" {
		t.Fatal("a plugin without the capability finishes its work")
	}
}

func TestShutdownAnswersWhatIsInFlightAndWaitingWithTheirResultsFirst(t *testing.T) {
	for name, p := range map[string]interface {
		serve.Plugin
		unlatch()
		waitStarted(*testing.T, int) []string
	}{"cancellable": cancellableLab{newLab(1)}, "plain": plainLab{newLab(1)}} {
		w := start(t, p)
		w.initialize()
		w.discover(1, "stubborn-running")
		w.discover(2, "stubborn-waiting")
		w.request(3, protocol.MethodShutdown, struct{}{})
		p.waitStarted(t, 1)
		w.silent(50 * time.Millisecond) // shutdown waits on the latch: nothing is cancelled, nothing answered
		p.unlatch()
		for _, want := range []int64{1, 2, 3} {
			resp := w.next()
			if *resp.ID != want || resp.Error != nil {
				t.Fatalf("%s: want a result for %d, got %+v", name, want, resp)
			}
		}
		select {
		case <-w.done:
			if w.serveErr != nil {
				t.Fatal(w.serveErr)
			}
		case <-time.After(bound):
			t.Fatalf("%s: Serve did not return after shutdown", name)
		}
	}
}

func TestTheEndOfTheInputCancelsWhatIsInFlightOfACancellablePlugin(t *testing.T) {
	p := cancellableLab{newLab(0)}
	w := start(t, p)
	w.initialize()
	w.discover(1, "block-a")
	p.waitStarted(t, 1)
	w.closeIn()
	select {
	case <-w.done:
	case <-time.After(bound):
		t.Fatal("Serve did not return after the host closed the stream")
	}
	if p.cancelled != 1 {
		t.Fatalf("cancelled %d", p.cancelled)
	}
}

// plainLab is a lab that does not advertise cancel (lab itself has no WithContext).
type plainLab struct{ lab }

func (l lab) unlatch() { close(l.release) }

func TestARequestIdAlreadyInFlightAndAFloodAreRefused(t *testing.T) {
	p := newLab(1)
	w := start(t, p)
	w.initialize()
	w.discover(1, "block-a")
	p.waitStarted(t, 1)
	w.discover(1, "block-again")
	if resp := w.next(); resp.Error == nil || resp.Error.Code != protocol.CodeInvalidRequest {
		t.Fatalf("a reused id is invalid: %+v", resp)
	}
	refused := 0
	for id := int64(2); id < 2+1024+10; id++ {
		w.discover(id, "block-queued")
	}
	for range 10 {
		resp := w.next()
		if resp.Error == nil || resp.Error.Code != protocol.CodeInvalidRequest || !strings.Contains(resp.Error.Message, "too many requests") {
			t.Fatalf("beyond the queue the host is refused: %+v", resp)
		}
		refused++
	}
	if refused != 10 {
		t.Fatalf("refused %d", refused)
	}
	close(p.release)
	for range 1 + 1024 {
		project(t, w.next())
	}
}

func TestRequestsBeforeInitializeAreRefused(t *testing.T) {
	w := start(t, newLab(0))
	w.discover(1, "quick")
	if resp := w.next(); resp.Error == nil || resp.Error.Code != protocol.CodeInvalidRequest {
		t.Fatalf("%+v", resp)
	}
}

func TestAStressOfConcurrentAndCancelledRequestsAnswersEachExactlyOnceWithinTheBound(t *testing.T) {
	const total, workers = 600, 8
	before := runtime.NumGoroutine()
	p := cancellableLab{newLab(workers)}
	w := start(t, p)
	w.initialize()
	for id := int64(1); id <= total; id++ {
		name := "slow-x"
		if id%3 == 0 {
			name = "block-x"
		}
		w.discover(id, name)
		if id%3 == 0 && id%2 == 0 {
			go func() {
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				w.notify(protocol.MethodCancel, protocol.CancelParams{ID: id})
			}()
		}
	}
	// Block requests nobody cancelled are released at the end.
	go func() { time.Sleep(200 * time.Millisecond); close(p.release) }()
	answered := map[int64]int{}
	for range total {
		resp := w.next()
		answered[*resp.ID]++
		if resp.Error != nil && resp.Error.Code != protocol.CodeRequestCancelled {
			t.Fatalf("unexpected error %+v", resp.Error)
		}
	}
	for id := int64(1); id <= total; id++ {
		if answered[id] != 1 {
			t.Fatalf("request %d answered %d times", id, answered[id])
		}
	}
	w.silent(30 * time.Millisecond)
	if peak := p.peakActive(); peak > workers {
		t.Fatalf("%d handlers at once, the pool is bounded to %d", peak, workers)
	}
	w.closeIn()
	select {
	case <-w.done:
		err := w.serveErr
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(bound):
		t.Fatal("Serve did not return")
	}
	time.Sleep(100 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+4 {
		t.Fatalf("goroutines leaked: %d before, %d after", before, after)
	}
}

func TestACancellableErrorIsStillAPluginFailureWhenNobodyCancelled(t *testing.T) {
	p := cancellableLab{newLab(0)}
	w := start(t, p)
	w.initialize()
	w.discover(1, "fail-by-itself")
	resp := w.next()
	if resp.Error == nil || resp.Error.Code != protocol.CodePluginFailure {
		t.Fatalf("an error that is not a cancel keeps its own code: %+v", resp)
	}
}
