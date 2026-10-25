package host_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/transport"
)

type fakePlugin struct{}

func (fakePlugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "fake", Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"fake.json"}}, nil
}

func (fakePlugin) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	var out protocol.DiscoverResult
	for _, m := range p.Manifests {
		if m == "bad/fake.json" {
			return out, errors.New("bad manifest")
		}
		out.Projects = append(out.Projects, protocol.Project{Name: m, Root: m, Kind: protocol.ProjectLib})
	}
	return out, nil
}

func connect(t *testing.T) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fakePlugin{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: "/ws"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestHandshakeAndCapabilities(t *testing.T) {
	c := connect(t)
	if c.Info.Plugin.Name != "fake" || c.Info.ProtocolVersion != protocol.Version {
		t.Fatalf("unexpected info %+v", c.Info)
	}
	if !c.Info.Capabilities.Discover || c.Info.Capabilities.Generate {
		t.Fatalf("capabilities must mirror implemented interfaces: %+v", c.Info.Capabilities)
	}
	if !c.Supports(protocol.MethodDiscover) || c.Supports(protocol.MethodPublish) {
		t.Fatal("Supports disagrees with capabilities")
	}
}

func TestCallsAndErrors(t *testing.T) {
	c := connect(t)
	var res protocol.DiscoverResult
	if err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{"a/fake.json"}}, &res, 0); err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 || res.Projects[0].Name != "a/fake.json" {
		t.Fatalf("unexpected result %+v", res)
	}
	err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{"bad/fake.json"}}, &res, 0)
	var rpcErr *protocol.ResponseError
	if !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodePluginFailure {
		t.Fatalf("want plugin failure, got %v", err)
	}
	err = c.Call(protocol.MethodPublish, struct{}{}, nil, 0)
	if !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodeMethodNotFound {
		t.Fatalf("want method not found, got %v", err)
	}
}

func TestRejectsUnknownProtocolVersion(t *testing.T) {
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fakePlugin{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var res protocol.InitializeResult
	err = c.Call(protocol.MethodInitialize, protocol.InitializeParams{ProtocolVersions: []int{99}}, &res, 0)
	if err == nil {
		t.Fatal("expected a version negotiation error")
	}
	_ = c.Close()
}

// state is what a lab plugin shares across the copies Serve makes of it.
type state struct {
	concurrency int
	started     chan string
	release     chan struct{}
	mu          sync.Mutex
	active      int
	peak        int
}

// lab is a plugin whose discover does what the first manifest name says:
// "block-*" waits to be released or, if the plugin is cancellable and listens,
// cancelled; "deaf-*" waits to be released and ignores cancellation; anything
// else answers at once.
type lab struct {
	*state
	ctx context.Context
}

func newLab(concurrency int) lab {
	return lab{state: &state{concurrency: concurrency, started: make(chan string, 256), release: make(chan struct{})}}
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
	var done <-chan struct{}
	if l.ctx != nil {
		done = l.ctx.Done()
	}
	switch {
	case strings.HasPrefix(name, "block"):
		select {
		case <-l.release:
		case <-done:
			return protocol.DiscoverResult{}, l.ctx.Err()
		}
	case strings.HasPrefix(name, "deaf"):
		<-l.release
	case strings.HasPrefix(name, "hold"):
		time.Sleep(20 * time.Millisecond)
	}
	return protocol.DiscoverResult{Projects: []protocol.Project{{Name: name, Root: name, Language: "lab", Kind: protocol.ProjectLib}}}, nil
}

func (l lab) peakActive() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.peak
}

// cancellableLab is a lab that advertises cancel.
type cancellableLab struct{ lab }

func (c cancellableLab) WithContext(ctx context.Context) serve.Plugin {
	c.ctx = ctx
	return c
}

// spy sits between the host and the plugin and records the methods the host
// sends.
type spy struct {
	mu      sync.Mutex
	methods []string
}

func (s *spy) saw(method string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, m := range s.methods {
		if m == method {
			n++
		}
	}
	return n
}

func connectSpied(t *testing.T, p serve.Plugin, opts host.Options) (*host.Client, *spy) {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, middleW := io.Pipe()
	hostRead, hostW := io.Pipe()
	s := &spy{}
	go func() {
		in, out := transport.NewStream(hostRead, nil), transport.NewStream(nil, middleW)
		for {
			body, err := in.Read()
			if err != nil {
				_ = middleW.Close()
				return
			}
			var req protocol.Request
			if json.Unmarshal(body, &req) == nil {
				s.mu.Lock()
				s.methods = append(s.methods, req.Method)
				s.mu.Unlock()
			}
			_ = out.Write(body)
		}
	}()
	go func() { _ = serve.Serve(p, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c, s
}

func discover(c *host.Client, ctx context.Context, name string, timeout time.Duration) (string, error) {
	var res protocol.DiscoverResult
	if err := c.CallContext(ctx, protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{name}}, &res, timeout); err != nil {
		return "", err
	}
	return res.Projects[0].Name, nil
}

func waitStarted(t *testing.T, l lab, n int) {
	t.Helper()
	for range n {
		select {
		case <-l.started:
		case <-time.After(5 * time.Second):
			t.Fatal("the plugin did not start the request")
		}
	}
}

func callTogether(t *testing.T, c *host.Client, n int, name string) {
	t.Helper()
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got, err := discover(c, context.Background(), name, 10*time.Second); err != nil || got != name {
				t.Errorf("discover: %q, %v", got, err)
			}
		}()
	}
	wg.Wait()
}

func TestAPluginThatDeclaresNothingIsCalledOneRequestAtATime(t *testing.T) {
	l := newLab(0)
	c, _ := connectSpied(t, l, host.Options{})
	callTogether(t, c, 8, "hold")
	if l.peakActive() != 1 || c.PeakInFlight() != 1 {
		t.Fatalf("the plugin saw %d requests at once and the host had %d in flight, want 1", l.peakActive(), c.PeakInFlight())
	}
}

func TestAPluginIsNeverCalledMoreOftenAtOnceThanItDeclared(t *testing.T) {
	l := newLab(3)
	c, _ := connectSpied(t, l, host.Options{})
	callTogether(t, c, 12, "hold")
	if l.peakActive() != 3 || c.PeakInFlight() != 3 {
		t.Fatalf("the plugin saw %d requests at once and the host had %d in flight, want 3", l.peakActive(), c.PeakInFlight())
	}
}

func TestAnUnlimitedPluginIsCalledUpToTheHostsOwnBound(t *testing.T) {
	l := newLab(protocol.ConcurrencyUnlimited)
	c, _ := connectSpied(t, l, host.Options{})
	callTogether(t, c, host.UnlimitedInFlight*2, "hold")
	if c.PeakInFlight() > host.UnlimitedInFlight || c.PeakInFlight() < 2 {
		t.Fatalf("the host had %d in flight, want between 2 and %d", c.PeakInFlight(), host.UnlimitedInFlight)
	}
}

func TestCancellingAContextStopsACancellablePluginAndKeepsItRunning(t *testing.T) {
	l := cancellableLab{newLab(0)}
	c, s := connectSpied(t, l, host.Options{})
	if !c.Supports(protocol.MethodCancel) {
		t.Fatal("the host must see the cancel capability")
	}
	ctx, cancel := context.WithCancel(context.Background())
	failed := make(chan error, 1)
	go func() { _, err := discover(c, ctx, "block-a", time.Minute); failed <- err }()
	waitStarted(t, l.lab, 1)
	began := time.Now()
	cancel()
	err := <-failed
	var rpcErr *protocol.ResponseError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodeRequestCancelled {
		t.Fatalf("want context.Canceled and the plugin's cancelled answer, got %v", err)
	}
	if time.Since(began) > 2*time.Second {
		t.Fatalf("the call took %s to return after the cancel", time.Since(began))
	}
	if s.saw(protocol.MethodCancel) != 1 {
		t.Fatalf("the plugin was sent %d cancel notifications, want 1", s.saw(protocol.MethodCancel))
	}
	if got, err := discover(c, context.Background(), "after", 0); err != nil || got != "after" {
		t.Fatalf("the plugin must stay usable after a cancel: %q %v", got, err)
	}
}

func TestATimeoutAsksACancellablePluginToStopBeforeAnythingElse(t *testing.T) {
	l := cancellableLab{newLab(0)}
	c, s := connectSpied(t, l, host.Options{})
	_, err := discover(c, context.Background(), "block-a", 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want a deadline error, got %v", err)
	}
	if s.saw(protocol.MethodCancel) != 1 {
		t.Fatalf("cancel notifications sent: %d", s.saw(protocol.MethodCancel))
	}
	if got, err := discover(c, context.Background(), "after", 0); err != nil || got != "after" {
		t.Fatalf("the plugin was not killed, it must still answer: %q %v", got, err)
	}
}

func TestAPluginThatIgnoresCancelIsKilledAfterTheGracePeriod(t *testing.T) {
	l := cancellableLab{newLab(0)}
	c, _ := connectSpied(t, l, host.Options{CancelGrace: 150 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	failed := make(chan error, 1)
	go func() { _, err := discover(c, ctx, "deaf-a", time.Minute); failed <- err }()
	waitStarted(t, l.lab, 1)
	began := time.Now()
	cancel()
	err := <-failed
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "killed") {
		t.Fatalf("want a cancelled-and-killed error, got %v", err)
	}
	if took := time.Since(began); took < 150*time.Millisecond || took > 3*time.Second {
		t.Fatalf("the host waited %s, want about the 150ms grace", took)
	}
	if _, err := discover(c, context.Background(), "after", time.Second); err == nil {
		t.Fatal("a killed plugin does not answer")
	}
	close(l.release)
}

func TestAResultThatArrivesWhileTheHostIsStoppingTheCallIsTheCallsResult(t *testing.T) {
	l := cancellableLab{newLab(0)}
	c, _ := connectSpied(t, l, host.Options{CancelGrace: 2 * time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	type answer struct {
		name string
		err  error
	}
	got := make(chan answer, 1)
	go func() { name, err := discover(c, ctx, "deaf-a", time.Minute); got <- answer{name, err} }()
	waitStarted(t, l.lab, 1)
	cancel()
	time.Sleep(50 * time.Millisecond)
	close(l.release)
	if a := <-got; a.err != nil || a.name != "deaf-a" {
		t.Fatalf("the work finished, so the call has its result: %+v", a)
	}
}

func TestAPluginThatCannotBeCancelledIsNeverSentCancel(t *testing.T) {
	l := newLab(0)
	c, s := connectSpied(t, l, host.Options{CancelGrace: 100 * time.Millisecond})
	if c.Supports(protocol.MethodCancel) {
		t.Fatal("a plain plugin does not support cancel")
	}
	ctx, cancel := context.WithCancel(context.Background())
	failed := make(chan error, 1)
	go func() { _, err := discover(c, ctx, "block-a", time.Minute); failed <- err }()
	waitStarted(t, l, 1)
	cancel()
	if err := <-failed; !errors.Is(err, context.Canceled) {
		t.Fatalf("want a cancelled error, got %v", err)
	}
	if s.saw(protocol.MethodCancel) != 0 {
		t.Fatal("cancel was sent to a plugin that did not advertise it")
	}
}

func TestATimeoutKillsAPluginThatCannotBeCancelledAtOnce(t *testing.T) {
	l := newLab(0)
	c, s := connectSpied(t, l, host.Options{CancelGrace: time.Minute})
	began := time.Now()
	_, err := discover(c, context.Background(), "block-a", 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "did not answer within") {
		t.Fatalf("got %v", err)
	}
	if time.Since(began) > 3*time.Second || s.saw(protocol.MethodCancel) != 0 {
		t.Fatalf("it takes the timeout, not the grace period, and sends no cancel (took %s)", time.Since(began))
	}
	close(l.release)
}

func TestAContextCancelledWhileWaitingForAPlaceNeverSendsTheRequest(t *testing.T) {
	l := newLab(0)
	c, _ := connectSpied(t, l, host.Options{})
	running := make(chan error, 1)
	go func() { _, err := discover(c, context.Background(), "block-a", time.Minute); running <- err }()
	waitStarted(t, l, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := discover(c, ctx, "block-b", time.Minute); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	select {
	case name := <-l.started:
		t.Fatalf("the waiting call reached the plugin: %s", name)
	case <-time.After(50 * time.Millisecond):
	}
	close(l.release)
	if err := <-running; err != nil {
		t.Fatal(err)
	}
}
