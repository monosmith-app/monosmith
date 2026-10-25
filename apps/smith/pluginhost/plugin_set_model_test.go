package pluginhost

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// waiting is a plugin whose discover returns only when the host gives up on
// it, and which can be cancelled.
type waiting struct{ ctx context.Context }

func (waiting) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "waiting", Version: "1.0.0"}, Kind: protocol.KindLanguage}, nil
}

func (w waiting) WithContext(ctx context.Context) serve.Plugin { w.ctx = ctx; return w }

func (w waiting) Discover(protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	<-w.ctx.Done()
	return protocol.DiscoverResult{}, w.ctx.Err()
}

func TestACancelledContextStopsACallToAPluginInTheSet(t *testing.T) {
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(waiting{}, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	set := &PluginSet{Clients: []*host.Client{c}}

	ok, err := set.CallContext(context.Background(), "waiting", protocol.MethodGenerate, struct{}{}, nil, time.Second)
	if ok || err != nil {
		t.Fatalf("a method the plugin did not advertise is not called: %v %v", ok, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	began := time.Now()
	ok, err = set.CallContext(ctx, "waiting", protocol.MethodDiscover, protocol.DiscoverParams{}, nil, time.Minute)
	if !ok || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want a deadline error from a call that was made, got %v %v", ok, err)
	}
	if time.Since(began) > 3*time.Second {
		t.Fatalf("the call outlived its context by %s", time.Since(began))
	}
	if _, err := set.CallContext(context.Background(), "ghost", protocol.MethodDiscover, nil, nil, 0); err == nil {
		t.Fatal("an unknown plugin is an error")
	}
}
