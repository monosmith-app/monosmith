package conformance

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// scheduled is a language that declares a concurrency. With stuck set, its
// discover takes seconds and ignores any cancel.
type scheduled struct {
	concurrency int
	stuck       bool
}

func (s scheduled) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "sched", Version: "1.0.0"}, Kind: protocol.KindLanguage, Roles: []protocol.Role{protocol.RoleLanguage}, Manifests: []string{"sched.txt"}, Concurrency: s.concurrency}, nil
}

func (s scheduled) Discover(protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	if s.stuck {
		time.Sleep(3 * time.Second)
	}
	return protocol.DiscoverResult{}, nil
}

// cancellingScheduled advertises cancel.
type cancellingScheduled struct{ scheduled }

func (c cancellingScheduled) WithContext(context.Context) serve.Plugin { return c }

func connectWith(t *testing.T, p serve.Plugin, opts host.Options) *host.Client {
	t.Helper()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(p, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func names(r Report) string {
	var out []string
	for _, res := range r.Results {
		out = append(out, res.Check)
	}
	return strings.Join(out, "|")
}

func TestADefaultPluginIsSentRequestsOneAtATime(t *testing.T) {
	r := Check(connectWith(t, scheduled{}, host.Options{}), Options{})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
	for _, want := range []string{"declares a concurrency the protocol defines", "answers 3 requests sent together, one at a time"} {
		if !strings.Contains(names(r), want) {
			t.Errorf("%q did not run: %s", want, names(r))
		}
	}
	if strings.Contains(names(r), "cancel") {
		t.Errorf("a plugin without cancel is not checked for it: %s", names(r))
	}
}

func TestADeclaredConcurrencyIsAcceptedAndCancelIsChecked(t *testing.T) {
	r := Check(connectWith(t, cancellingScheduled{scheduled{concurrency: 4}}, host.Options{}), Options{})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
	for _, want := range []string{"accepts 4 requests at the same time", "answers a cancelled request with its result or the cancelled error", "ignores a cancel for a request it does not know"} {
		if !strings.Contains(names(r), want) {
			t.Errorf("%q did not run: %s", want, names(r))
		}
	}
}

func TestAnUndefinedConcurrencyIsReported(t *testing.T) {
	failures := Check(connectWith(t, scheduled{concurrency: -5}, host.Options{}), Options{}).Failures()
	if !strings.Contains(failures, "declares a concurrency the protocol defines") {
		t.Errorf("not reported:\n%s", failures)
	}
}

func TestAPluginThatAdvertisesCancelAndIgnoresItIsReported(t *testing.T) {
	old := cancelBound
	cancelBound = 300 * time.Millisecond
	t.Cleanup(func() { cancelBound = old })
	c := connectWith(t, cancellingScheduled{scheduled{stuck: true}}, host.Options{CancelGrace: 200 * time.Millisecond})
	probe, ok := probeOf(c)
	if !ok {
		t.Fatal("no probe")
	}
	var r Report
	checkCancel(c, probe, &r)
	if r.OK() || !strings.Contains(r.Failures(), "answers a cancelled request with its result or the cancelled error") {
		t.Errorf("a plugin that does not stop is not reported: %+v", r.Results)
	}
}
