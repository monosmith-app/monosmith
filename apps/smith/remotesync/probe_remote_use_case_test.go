package remotesync

import (
	"context"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

func TestProbeReadsTheAnswerAsAState(t *testing.T) {
	fake, srv := newFakeStore(t)
	p := plan(t, srv)
	if got := Probe(context.Background(), p); got.State != StateConnected {
		t.Errorf("token accepted: %+v", got)
	}
	fake.token = "another-token"
	if got := Probe(context.Background(), p); got.State != StateRefused || got.Detail == "" {
		t.Errorf("token refused: %+v", got)
	}
	if got := Probe(context.Background(), session.RemotePlan{}); got.State != StateNone {
		t.Errorf("no remote: %+v", got)
	}
	if got := Probe(context.Background(), session.RemotePlan{URL: p.URL}); got.State != StateRefused {
		t.Errorf("no SMITH_CACHE_KEY: %+v", got)
	}
}

func TestProbeGivesUpOnARemoteThatDoesNotAnswer(t *testing.T) {
	fake, srv := newFakeStore(t)
	fake.delay = 300 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if got := Probe(ctx, plan(t, srv)); got.State != StateUnreachable {
		t.Errorf("slow remote: %+v", got)
	}
	closed := plan(t, srv)
	srv.Close()
	if got := Probe(context.Background(), closed); got.State != StateUnreachable {
		t.Errorf("closed remote: %+v", got)
	}
}

func TestStoreNameKeepsNoSecret(t *testing.T) {
	for in, want := range map[string]string{
		"s3://bucket/team/cache":                "s3://bucket/team/cache",
		"https://user:pass@cache.example.com/x": "https://cache.example.com/x",
		"azblob://acct/container?sig=secret":    "azblob://acct/container",
	} {
		if got := StoreName(in); got != want {
			t.Errorf("StoreName(%q) = %q, want %q", in, got, want)
		}
	}
}
